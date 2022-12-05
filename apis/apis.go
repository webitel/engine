package apis

import (
	"github.com/gorilla/mux"
	"github.com/webitel/engine/app"
)

type Routes struct {
	Root     *mux.Router // ''
	Endpoint *mux.Router // '/endpoint'
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
	api.Routes.Endpoint = api.Routes.Root.PathPrefix("/endpoint").Subrouter()

	api.InitWebSocket()
	api.InitAppointments()
	api.InitOAuth()

	return api
}
