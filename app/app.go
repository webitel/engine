package app

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/webitel/call_center/grpc_api/client"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/b2bua"
	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/chat_manager"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/logger"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/engine/mq/rabbit"
	"github.com/webitel/engine/presign"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/store/sqlstore"
	flow "github.com/webitel/flow_manager/client"
	otelsdk "github.com/webitel/webitel-go-kit/otel/sdk"
	"github.com/webitel/wlog"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/atomic"
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
	cc               client.CCManager
	cipher           presign.PreSign
	audit            *logger.Audit
	b2b              *b2bua.B2B
	ctx              context.Context
	tracer           *Tracer
	otelShutdownFunc otelsdk.ShutdownFunc
	TriggerCases     TriggerCase
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

	if localization.T == nil {
		if err := localization.TranslationsPreInit(config.TranslationsDirectory); err != nil {
			return nil, errors.Wrapf(err, "unable to load translation files")
		}
	}
	model.AppErrorInit(localization.T)

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

	if err := app.setupCipher(); err != nil {
		return nil, err
	}

	if err := localization.InitTranslations(model.LocalizationSettings{
		DefaultClientLocale: model.NewString(model.DEFAULT_LOCALE),
		DefaultServerLocale: model.NewString(model.DEFAULT_LOCALE),
		AvailableLocales:    model.NewString(model.DEFAULT_LOCALE),
	}); err != nil {
		return nil, errors.Wrapf(err, "unable to load translation files")
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

	app.Store = store.NewLayeredStore(sqlstore.NewSqlSupplier(app.Config().SqlSettings))

	app.MessageQueue = rabbit.NewRabbitMQ(app.Config().NodeName, &app.Config().MessageQueueSettings)
	app.MessageQueue.Start()

	app.Hubs = NewHubs(app)

	app.GrpcServer = NewGrpcServer(app, app.Config().ServerSettings)

	if outErr = app.cluster.Start(); outErr != nil {
		return nil, outErr
	}

	app.sessionManager = auth_manager.NewAuthManager(model.SESSION_CACHE_SIZE, app.Config().AuthCacheExpire, app.cluster.discovery, app.Log)
	if err := app.sessionManager.Start(); err != nil {
		return nil, err
	}

	app.chatManager = chat_manager.NewChatManager(app.cluster.discovery)
	if err := app.chatManager.Start(); err != nil {
		return nil, err
	}

	app.callManager = call_manager.NewCallManager(app.Config().SipSettings.ServerAddr, app.Config().SipSettings.Proxy, app.cluster.discovery)
	if err := app.callManager.Start(); err != nil {
		return nil, err
	}

	app.flowManager = flow.NewFlowManager(app.cluster.discovery)
	if err := app.flowManager.Start(); err != nil {
		return nil, err
	}

	app.cc = client.NewCCManager(app.cluster.discovery)
	if err := app.cc.Start(); err != nil {
		return nil, err
	}

	if app.audit, err = logger.New(app.Config().DiscoverySettings.Url, app.MessageQueue); err != nil {
		return nil, err
	}

	if config.B2BSettings.Addr != "" {
		app.b2b = b2bua.New(app, b2bua.Config{
			Addr:     config.B2BSettings.Addr,
			SipProxy: config.SipSettings.Proxy,
		})

	}

	// start triggers for cases
	if app.config.CaseTriggersSettings.Enabled {
		app.TriggerCases = NewTriggerCases(app.Log, app.Store, app.flowManager, &app.config.CaseTriggersSettings)
		if err := app.TriggerCases.Start(); err != nil {
			return nil, fmt.Errorf("unable to start cases trigger: %w", err)
		}
	}

	return app, outErr
}

func (app *App) Shutdown() {
	wlog.Info("stopping Server...")

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
	if app.TriggerCases != nil {
		app.TriggerCases.Stop()
	}

	app.cluster.Stop()
	app.sessionManager.Stop()
}

func (app *App) CallManager() call_manager.CallManager {
	return app.callManager
}

func (app *App) Ready() (bool, model.AppError) {
	//TODO
	return true, nil
}

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
