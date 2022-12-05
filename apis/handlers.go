package apis

import (
	"github.com/webitel/engine/web"
	"net/http"
)

type Context = web.Context

func (api *API) ApiHandlerTrustRequester(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	handler := &web.Handler{
		App:            api.App,
		HandleFunc:     h,
		RequireSession: false,
		TrustRequester: true,
		RequireMfa:     false,
		IsStatic:       false,
	}

	return handler
}

func (api *API) ApiHandlerRequester(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	handler := &web.Handler{
		App:            api.App,
		HandleFunc:     h,
		RequireSession: true,
		TrustRequester: true,
		RequireMfa:     false,
		IsStatic:       false,
	}

	return handler
}
