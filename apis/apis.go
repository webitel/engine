package apis

import (
	"github.com/gorilla/mux"
	"github.com/webitel/engine/app"
)

type Routes struct {
	Root *mux.Router // ''
}

type API struct {
	App    *app.App
	Routes *Routes
}

func Init(a *app.App, root *mux.Router) *API {
	api := &API{
		App:    a,
		Routes: &Routes{},
	}

	api.Routes.Root = root

	api.InitWebSocket()

	return api
}
