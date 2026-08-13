package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/atomic"

	"github.com/webitel/webitel-go-kit/infra/health"
	healthhttp "github.com/webitel/webitel-go-kit/infra/health/http"
	"github.com/webitel/webitel-go-kit/infra/health/sdnotify"
	otelsdk "github.com/webitel/webitel-go-kit/otel/sdk"
	"github.com/webitel/wlog"

	"github.com/webitel/engine/app/cc"
	"github.com/webitel/engine/app/flow"
	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/logger"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/engine/mq/rabbit"
	"github.com/webitel/engine/pkg/presign"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/pkg/wbt/chat_manager"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/store/sqlstore"
	"github.com/webitel/engine/wlogslog"

	// -------------------- plugin(s) -------------------- //
	_ "github.com/webitel/webitel-go-kit/otel/sdk/log/otlp"
	_ "github.com/webitel/webitel-go-kit/otel/sdk/log/stdout"
	_ "github.com/webitel/webitel-go-kit/otel/sdk/metric/otlp"
	_ "github.com/webitel/webitel-go-kit/otel/sdk/metric/stdout"
	_ "github.com/webitel/webitel-go-kit/otel/sdk/trace/otlp"
	_ "github.com/webitel/webitel-go-kit/otel/sdk/trace/stdout"
)

const (
	EventUpdateAction = "update"
	EventDeleteAction = "delete"
	EventCreateAction = "create"
	EventExchangeName = "event"
)

type App struct {
	nodeId           string
	config           *model.Config
	Log              *wlog.Logger
	Srv              *Server
	GrpcServer       *GrpcServer
	Hubs             *Hubs
	MessageQueue     mq.MQ
	Count            atomic.Int64
	Store            store.Store
	cluster          *cluster
	sessionManager   auth_manager.AuthManager
	callManager      call_manager.CallManager
	chatManager      chat_manager.ChatManager
	flowManager      flow.FlowManager
	cc               cc.CCManager
	cipher           presign.PreSign
	audit            *logger.Audit
	ctx              context.Context
	tracer           *Tracer
	otelShutdownFunc otelsdk.ShutdownFunc
	eventTrigger     EventTrigger
	health           *health.Registry
	sdNotify         *sdnotify.Notifier
}

func New(options ...string) (outApp *App, outErr error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	if err := config.IsValid(); err != nil {
		return nil, err
	}

	app := &App{
		nodeId: config.NodeName,
		config: config,
		Srv: &Server{
			RootRouter: mux.NewRouter(),
		},
		ctx: context.Background(),
	}

	app.Srv.Router = app.Srv.RootRouter.PathPrefix("/").Subrouter()

	setupPublicStorageUrl(config.PublicHostName)

	logConfig := &wlog.LoggerConfiguration{
		EnableConsole: config.Log.Console,
		ConsoleJson:   false,
		ConsoleLevel:  config.Log.Lvl,
	}

	if config.Log.File != "" {
		logConfig.FileLocation = config.Log.File
		logConfig.EnableFile = true
		logConfig.FileJson = true
		logConfig.FileLevel = config.Log.Lvl
	}

	if config.Log.Otel {
		// TODO
		logConfig.EnableExport = true
		app.otelShutdownFunc, err = otelsdk.Configure(
			app.ctx,
			otelsdk.WithResource(resource.NewSchemaless(
				semconv.ServiceName(model.APP_SERVICE_NAME),
				semconv.ServiceVersion(model.CurrentVersion),
				semconv.ServiceInstanceID(app.nodeId),
				semconv.ServiceNamespace("webitel"),
			)),
		)
		if err != nil {
			return nil, err
		}
	}
	app.tracer = NewTrace()

	app.Log = wlog.NewLogger(logConfig)

	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	// Health starts here, before anything slow. systemd counts
	// TimeoutStartSec from ExecStart, so WithStartTimeout has to be measured
	// from about the same moment — start the notifier after the managers and
	// its fallback READY=1 can land after systemd has already given up. The
	// registry reports not-ready until checks are registered further down,
	// which is what a booting node should say.
	healthLog := slog.New(wlogslog.NewHandler(app.Log))
	app.health = health.New(health.DefaultConfig(), healthLog)

	// app.ctx: Start's context is the scheduler's lifetime, so a short-lived
	// one would silently stop every check.
	if err := app.health.Start(app.ctx); err != nil {
		return nil, fmt.Errorf("unable to start health registry: %w", err)
	}

	// 60s leaves 30s of margin under the unit's TimeoutStartSec=90.
	// nil when NOTIFY_SOCKET is unset; Start and Stop are both nil-safe.
	app.sdNotify = sdnotify.New(app.health,
		sdnotify.WithLogger(healthLog),
		sdnotify.WithStartTimeout(60*time.Second),
	)
	if err := app.sdNotify.Start(app.ctx); err != nil {
		return nil, fmt.Errorf("unable to start sd_notify: %w", err)
	}

	// RootRouter, not the API subrouter: probes answer without a token.
	app.Srv.RootRouter.Handle("/livez", healthhttp.LivenessHandler(app.health, healthhttp.WithLogger(healthLog)))
	app.Srv.RootRouter.Handle("/readyz", healthhttp.ReadinessHandler(app.health, healthhttp.WithLogger(healthLog)))
	app.Srv.RootRouter.Handle("/healthz", healthhttp.HealthHandler(app.health, healthhttp.WithLogger(healthLog)))

	if err := app.setupCipher(); err != nil {
		return nil, err
	}

	app.cluster = NewCluster(app)

	if config.Push.FirebaseServiceAccount != "" {
		err = initFirebase(config.Push.FirebaseServiceAccount)
		if err != nil {
			return nil, err
		}
		wlog.Info("enable push firebase")
	} else {
		wlog.Info("disabled push firebase")
	}

	if config.Push.ValidApn() {
		err = initApn(config.Push)
		if err != nil {
			return nil, err
		}
		wlog.Info("enable push apn")
	} else {
		wlog.Info("disabled push apn")
	}

	app.Srv.WebSocketRouter = &WebSocketRouter{
		app:      app,
		handlers: make(map[string]webSocketHandler),
	}

	if config.RTCConfiguration != "" {
		err = InitRTCConfiguration([]byte(config.RTCConfiguration))
		if err != nil {
			return nil, err
		}
	}

	// Concrete handle: store.Store does not expose GetMaster.
	sqlSupplier := sqlstore.NewSqlSupplier(app.Config().SqlSettings)
	app.Store = store.NewLayeredStore(sqlSupplier)

	app.MessageQueue = rabbit.NewRabbitMQ(app.Config().NodeName, &app.Config().MessageQueueSettings)
	app.MessageQueue.Start()

	app.Hubs = NewHubs(app)
	// remove all session by app id
	app.Hubs.Clean()

	app.GrpcServer = NewGrpcServer(app, app.Config().ServerSettings)

	if outErr = app.cluster.Start(); outErr != nil {
		return nil, outErr
	}

	app.sessionManager = auth_manager.NewAuthManager(model.SESSION_CACHE_SIZE, app.Config().AuthCacheExpire,
		app.Config().DiscoverySettings.Url, app.Log)
	if err := app.sessionManager.Start(); err != nil {
		return nil, err
	}

	app.chatManager = chat_manager.NewChatManager(app.Config().DiscoverySettings.Url)
	if err := app.chatManager.Start(); err != nil {
		return nil, err
	}

	app.callManager = call_manager.NewCallManager(app.Config().SipSettings.ServerAddr, app.Config().SipSettings.Proxy, app.cluster.discovery)
	if err := app.callManager.Start(); err != nil {
		return nil, err
	}

	app.flowManager = flow.NewFlowManager(app.Config().SipSettings.ServerAddr)
	if err := app.flowManager.Start(); err != nil {
		return nil, err
	}

	app.cc = cc.NewCCManager(app.Config().DiscoverySettings.Url)
	if err := app.cc.Start(); err != nil {
		return nil, err
	}

	if app.audit, err = logger.New(app.MessageQueue); err != nil {
		return nil, err
	}

	// start triggers for cases
	if app.config.TriggersSettings.Enabled {
		app.eventTrigger = NewEventTrigger(app.Log, app.Store, app.flowManager, &app.config.TriggersSettings)
		if err := app.eventTrigger.Start(); err != nil {
			return nil, fmt.Errorf("unable to start cases trigger: %w", err)
		}
	}

	// Critical is for node-local faults only: a shared dependency marked
	// critical would take the whole fleet out of rotation at once. Consul is
	// deliberately unchecked — the verdict travels through it.
	app.health.Critical("grpc", health.ListenerCheck(app.GrpcServer.Listener()))
	app.health.Critical("freeswitch", freeswitchCheck(app.callManager))
	app.health.Informational("postgres", func(ctx context.Context) error {
		return sqlSupplier.GetMaster().Db.PingContext(ctx)
	})

	if p, ok := app.MessageQueue.(interface {
		Ping(context.Context) error
	}); ok {
		app.health.Informational("rabbitmq", p.Ping)
	}

	return app, outErr
}

func (app *App) Shutdown() {
	wlog.Info("stopping Server...")

	// Drain before anything is torn down, so the node stops advertising
	// readiness while its dependencies are still up. The DrainHold wait happens
	// inside Stop: 12s clears the 10s hold and fits TimeoutStopSec=30. Stop also
	// halts the scheduler before MessageQueue.Close, so the rabbitmq check
	// cannot race a closing connection.
	if app.health != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		if err := health.Shutdown(ctx, app.health, app.sdNotify); err != nil {
			wlog.Error(fmt.Sprintf("health shutdown: %s", err.Error()))
		}

		cancel()
	}

	if app.Hubs != nil {
		app.Hubs.Clean()
	}

	if app.MessageQueue != nil {
		app.MessageQueue.Close()
	}

	if app.GrpcServer != nil {
		app.GrpcServer.srv.Stop()
	}

	if app.callManager != nil {
		app.callManager.Stop()
	}

	if app.chatManager != nil {
		app.chatManager.Stop()
	}

	if app.flowManager != nil {
		app.flowManager.Stop()
	}

	if app.otelShutdownFunc != nil {
		app.otelShutdownFunc(app.ctx)
	}

	// shutdown Cases Triggers
	if app.eventTrigger != nil {
		app.eventTrigger.Stop()
	}

	app.cluster.Stop()
	app.sessionManager.Stop()
}

func (app *App) CallManager() call_manager.CallManager {
	return app.callManager
}

// Ready reports whether this node can take traffic, per the health registry.
func (app *App) Ready() (bool, model.AppError) {
	if app.health == nil {
		return false, model.NewInternalError("app.ready.no_registry", "health registry is not initialized")
	}

	ok, err := app.health.ReadyFunc()()
	if ok {
		return true, nil
	}

	reason := "not ready"
	if err != nil {
		reason = err.Error()
	}

	return false, model.NewInternalError("app.ready.not_ready", reason)
}

// DEPRECATED use SendDomainEvent instead
func (a *App) PublishEventContext(ctx context.Context, body []byte, object string, keys ...string) model.AppError {
	routingKey := object
	for _, key := range keys {
		routingKey += fmt.Sprintf(".%s", key)
	}
	err := a.MessageQueue.Send(ctx, EventExchangeName, routingKey, body)
	if err != nil {
		return model.NewInternalError("app.app.publish_event_context.send.error", err.Error())
	}
	return nil
}

type DomainEventType string

const (
	CreateType DomainEventType = "create"
	DeleteType DomainEventType = "delete"
	UpdateType DomainEventType = "update"
)

type DomainEvent struct {
	DomainID  int64
	Object    string
	EventType DomainEventType
	User      int64
	Time      time.Time
	Body      any
}

func (d *DomainEvent) Validate() error {
	if d == nil {
		return errors.New("domain event is nil")
	}
	if d.DomainID <= 0 {
		return errors.New("domain id required for the domain event")
	}
	if d.Object == "" {
		return errors.New("object required for the domain event")
	}
	if d.EventType == "" {
		return errors.New("event type is required for the domain event")
	}
	return nil
}

func formatDomainEventKey(event *DomainEvent) (string, error) {
	if event.Object == "" {
		return "", errors.New("object required")
	}

	if event.EventType == "" {
		return "", errors.New("event type required")
	}
	return fmt.Sprintf("%s.%s.%d", event.Object, event.EventType, event.DomainID), nil
}

func (a *App) SendDomainEvent(ctx context.Context, event *DomainEvent) error {
	err := event.Validate()
	if err != nil {
		return err
	}
	routingKey, err := formatDomainEventKey(event)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event.Body)
	if err != nil {
		return err
	}

	return a.MessageQueue.Send(ctx, EventExchangeName, routingKey, body)
}
