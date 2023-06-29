package apis

import (
	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
	"net/http"
)

func (api *API) InitOAuth() {
	api.Routes.Endpoint.Handle("/oauth2/{id}/callback", api.ApiHandlerTrustRequester(handleOAuth2Callback)).Methods("GET")
}

func handleOAuth2Callback(c *Context, w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	props := mux.Vars(r)
	e, ok := c.App.MailOauthConfig(props["id"])
	if !ok {
		c.Err = model.NewBadRequestError("api.oauth2.callback.bad_request", "Not found provider "+props["id"])
		return
	}

	//oauthState, _ := r.Cookie("oauthstate")
	//
	//if r.FormValue("state") != oauthState.Value {
	//	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	//	return
	//}

	id, err := c.App.DecryptId(r.FormValue("state"))
	if err != nil {
		c.Err = err
		return
	}

	token, err2 := e.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err2 != nil {
		c.Err = model.NewBadRequestError("api.oauth2.callback.bad_request", err2.Error())
		return
	}

	if c.Err = c.App.EmailLoginOAuth(r.Context(), int(id), token); c.Err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}
