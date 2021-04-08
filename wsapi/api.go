package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/controller"
)

type API struct {
	App    *app.App
	ctrl   *controller.Controller
	Router *app.WebSocketRouter
}

func Init(a *app.App, router *app.WebSocketRouter) {
	api := &API{
		App:    a,
		Router: router,
		ctrl:   controller.NewController(a),
	}

	api.InitUser()
	api.InitCall()
	api.InitAgent()
	api.InitMember()
	api.InitChat()
	api.InitNotification()
}
