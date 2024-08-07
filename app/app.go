package app

import (
	"context"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/webitel/call_center/grpc_api/client"
	"github.com/webitel/webitel-go-kit/logging"
	"github.com/webitel/webitel-go-kit/tracing"
	"github.com/webitel/wlog"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"
	"gopkg.in/natefinch/lumberjack.v2"

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
)

const (
	EventUpdateAction = "update"
	EventDeleteAction = "delete"
	EventCreateAction = "create"
	EventExchangeName = "event"
)

type App struct {
	nodeId          string
	config          *model.Config
	Log             *wlog.Logger
	Srv             *Server
	GrpcServer      *GrpcServer
	Hubs            *Hubs
	MessageQueue    mq.MQ
	Count           atomic.Int64
	Store           store.Store
	cluster         *cluster
	sessionManager  auth_manager.AuthManager
	callManager     call_manager.CallManager
	chatManager     chat_manager.ChatManager
	cc              client.CCManager
	cipher          presign.PreSign
	audit           *logger.Audit
	b2b             *b2bua.B2B
	loggingProvider logging.LoggerProvider
	tracingProvider tracing.TracerProvider
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
	}

	app.Srv.Router = app.Srv.RootRouter.PathPrefix("/").Subrouter()

	setupPublicStorageUrl(config.PublicHostName)

	if localization.T == nil {
		if err := localization.TranslationsPreInit(config.TranslationsDirectory); err != nil {
			return nil, errors.Wrapf(err, "unable to load translation files")
		}
	}
	model.AppErrorInit(localization.T)

	initOTLPExporter := func(ctx context.Context, opts ...logging.Option) error {
		opts = append(opts, logging.WithServiceVersion(model.CurrentVersion),
			logging.WithAttributes(attribute.String("service.id", config.NodeName),
				attribute.Int("service.build", model.BuildNumberInt()),
			),
		)

		app.loggingProvider, err = logging.New(context.Background(), model.APP_SERVICE_NAME, opts...)
		if err != nil {
			return fmt.Errorf("unable to initialize logger provider: %v", err)
		}

		return nil
	}

	var logCfg *wlog.LoggerConfiguration
	switch config.Log.Format {
	case "legacy":
		logCfg = &wlog.LoggerConfiguration{
			EnableConsole: true,
			ConsoleLevel:  config.Log.Lvl,
		}
	case "legacy-json":
		logCfg = &wlog.LoggerConfiguration{}
		if config.Log.File != "" {
			logCfg.EnableFile = true
			logCfg.FileLevel = config.Log.Lvl
			logCfg.FileJson = true
			logCfg.FileLocation = config.Log.File
		} else {
			logCfg.EnableConsole = true
			logCfg.ConsoleLevel = config.Log.Lvl
			logCfg.ConsoleJson = true
		}

	case "otlp":
		logCfg = &wlog.LoggerConfiguration{
			EnableExport: true,
		}

		var lo []logging.Option
		if config.Log.Exporter == "" {
			return nil, fmt.Errorf("log exporter address must be configured if you wish to send logs to OTLP-compatible endpoint")
		}

		lo = append(lo, logging.WithExporter(string(logging.OTLPExporter)),
			logging.WithAddress(config.Log.Exporter),
		)

		if err = initOTLPExporter(context.Background(), lo...); err != nil {
			return nil, err
		}
	case "otlp-file":
		logCfg = &wlog.LoggerConfiguration{
			EnableExport: true,
		}

		if config.Log.File == "" {
			return nil, fmt.Errorf("log file location must be configured if you wish to store logs in a file")
		}

		w := &lumberjack.Logger{
			Filename: config.Log.File,
			Compress: true,
		}

		var lo []logging.Option
		lo = append(lo, logging.WithExporter(string(logging.STDOutExporter)),
			logging.WithSTDOutWriter(w),
		)

		if err = initOTLPExporter(context.Background(), lo...); err != nil {
			return nil, err
		}
	}

	app.Log = wlog.NewLogger(logCfg)
	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	app.tracingProvider, err = tracing.New(app.Log, model.APP_SERVICE_NAME,
		tracing.WithServiceVersion(model.CurrentVersion),
		tracing.WithAttributes(attribute.String("service.id", config.NodeName),
			attribute.Int("service.build", model.BuildNumberInt()),
		),
	)
	if err != nil {
		return nil, err
	}

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

	app.GrpcServer = NewGrpcServer(app.Config().ServerSettings)

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

	app.cluster.Stop()
	app.sessionManager.Stop()

	if app.loggingProvider != nil {
		app.loggingProvider.Shutdown(context.Background())
	}

	if app.tracingProvider != nil {
		app.tracingProvider.Shutdown(context.Background())
	}
}

func (app *App) CallManager() call_manager.CallManager {
	return app.callManager
}

func (app *App) Ready() (bool, model.AppError) {
	// TODO
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
