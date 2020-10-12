package app

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/webitel/call_center/grpc_api/client"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/chat_manager"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/engine/mq/rabbit"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/store/sqlstore"
	"github.com/webitel/wlog"
	"go.uber.org/atomic"
)

type App struct {
	nodeId         string
	config         *model.Config
	Log            *wlog.Logger
	Srv            *Server
	GrpcServer     *GrpcServer
	Hubs           *Hubs
	MessageQueue   mq.MQ
	Count          atomic.Int64
	Store          store.Store
	cluster        *cluster
	sessionManager auth_manager.AuthManager
	callManager    call_manager.CallManager
	chatManager    chat_manager.ChatManager
	cc             client.CCManager
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

	if localization.T == nil {
		if err := localization.TranslationsPreInit(config.TranslationsDirectory); err != nil {
			return nil, errors.Wrapf(err, "unable to load translation files")
		}
	}
	model.AppErrorInit(localization.T)

	app.Log = wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleLevel:  wlog.LevelDebug,
	})

	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	if err := localization.InitTranslations(model.LocalizationSettings{
		DefaultClientLocale: model.NewString(model.DEFAULT_LOCALE),
		DefaultServerLocale: model.NewString(model.DEFAULT_LOCALE),
		AvailableLocales:    model.NewString(model.DEFAULT_LOCALE),
	}); err != nil {
		return nil, errors.Wrapf(err, "unable to load translation files")
	}

	app.cluster = NewCluster(app)

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

	app.sessionManager = auth_manager.NewAuthManager(model.SESSION_CACHE_SIZE, model.SESSION_CACHE_TIME, app.cluster.discovery)
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
}

func (app *App) CallManager() call_manager.CallManager {
	return app.callManager
}

func (app *App) Ready() (bool, *model.AppError) {
	//TODO
	return true, nil
}
