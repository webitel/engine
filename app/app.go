// Package App Engine API.
//
// the purpose of this application is to provide an application
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /v2
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: John Doe<john.doe@example.com>
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - X-Webitel-Access:
//
//     SecurityDefinitions:
//     X-Webitel-Access:
//          type: apiKey
//          name: KEY
//          in: header
//
//     Extensions:
//     x-meta-value: value
//
// swagger:meta
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

	app.Store = store.NewLayeredStore(sqlstore.NewSqlSupplier(app.Config().SqlSettings))

	app.MessageQueue = rabbit.NewRabbitMQ(app.Config().NodeName, &app.Config().MessageQueueSettings)
	app.MessageQueue.Start()

	app.Hubs = NewHubs(app)

	app.GrpcServer = NewGrpcServer(app.Config().ServerSettings)

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

	if app.GrpcServer != nil {
		app.GrpcServer.srv.Stop()
	}

	app.cluster.Stop()
	app.sessionManager.Stop()
}

func (app *App) Ready() (bool, *model.AppError) {
	//TODO
	return true, nil
}
