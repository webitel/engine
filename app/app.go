package app

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/engine/mq/rabbit"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/store/sqlstore"
	"github.com/webitel/engine/utils"
	"github.com/webitel/wlog"
	"go.uber.org/atomic"
)

const (
	MAX_SESSION_SIZE   = 10000
	EXPIRE_SESSION_SEC = 60 * 5
)

type App struct {
	nodeId         string
	Log            *wlog.Logger
	Srv            *Server
	GrpcServer     *GrpcServer
	Hubs           *Hubs
	MessageQueue   mq.MQ
	Count          atomic.Int64
	Store          store.Store
	cluster        *cluster
	sessionManager auth_manager.AuthManager
}

func New(options ...string) (outApp *App, outErr error) {
	app := &App{
		nodeId: "engine-1",
		Srv: &Server{
			RootRouter: mux.NewRouter(),
		},
	}

	app.Srv.Router = app.Srv.RootRouter.PathPrefix("/").Subrouter()

	if utils.T == nil {
		if err := utils.TranslationsPreInit(); err != nil {
			return nil, errors.Wrapf(err, "unable to load translation files")
		}
	}
	model.AppErrorInit(utils.T)

	app.Log = wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleLevel:  wlog.LevelDebug,
	})

	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	if err := utils.InitTranslations(model.LocalizationSettings{
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

	app.Store = store.NewLayeredStore(sqlstore.NewSqlSupplier(model.SqlSettings{
		DriverName:                  model.NewString("postgres"),
		DataSource:                  model.NewString("postgres://webitel:webitel@localhost:5432/webitel?fallback_application_name=engine&sslmode=disable&connect_timeout=10&search_path=call_center"),
		MaxIdleConns:                model.NewInt(5),
		MaxOpenConns:                model.NewInt(5),
		ConnMaxLifetimeMilliseconds: model.NewInt(3600000),
		Trace:                       true,
	}))

	app.MessageQueue = rabbit.NewRabbitMQ("todo", &model.MessageQueueSettings{
		//Url: "amqp://webitel:webitel@10.10.10.200:5672?heartbeat=0",
		Url: "amqp://webitel:webitel@192.168.177.10:5672?heartbeat=10",
	})
	app.MessageQueue.Start()

	app.Hubs = NewHubs(app)

	app.GrpcServer = NewGrpcServer()

	if outErr = app.cluster.Start(); outErr != nil {
		return nil, outErr
	}

	app.sessionManager = auth_manager.NewAuthManager(app.cluster.discovery)
	if err := app.sessionManager.Start(); err != nil {
		return nil, err
	}

	return app, outErr
}

func (app *App) Shutdown() {
	wlog.Info("stopping Server...")

	if app.MessageQueue != nil {
		app.MessageQueue.Close()
	}

	app.cluster.Stop()
	app.sessionManager.Stop()
}

func (app *App) Ready() (bool, *model.AppError) {
	//TODO
	return true, nil
}
