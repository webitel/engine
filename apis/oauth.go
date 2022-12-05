package apis

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

func (api *API) InitOAuth() {
	api.Routes.Endpoint.Handle("/oauth2/{id}/callback", api.ApiHandlerTrustRequester(handleOAuth2Callback)).Methods("GET")
	api.Routes.Endpoint.Handle("/oauth2/{id}/login", api.ApiHandlerTrustRequester(handleOAuth2Login)).Methods("GET")
}

func handleOAuth2Callback(c *Context, w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	props := mux.Vars(r)
	e, ok := c.App.Config().EmailOAuth[props["id"]]
	if !ok {
		c.Err = model.NewAppError("API", "api.oauth2.callback.bad_request", nil, "Not found provider "+props["id"], http.StatusBadRequest)
		return
	}

	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	id, err := app.DecryptId(oauthState.Value)
	if err != nil {
		c.Err = err
		return
	}

	token, _ := e.Exchange(oauth2.NoContext, r.FormValue("code"))

	if c.Err = c.App.EmailLoginOAuth(int(id), token); c.Err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}
func handleOAuth2Login(c *Context, w http.ResponseWriter, r *http.Request) {
	props := mux.Vars(r)
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w, 77) // TODO
	e, ok := c.App.Config().EmailOAuth[props["id"]]
	if !ok {

	}

	//u := e.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("approval_prompt", "force"))
	u := e.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("approval_prompt", "force"))
	fmt.Println(u)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter, id int64) string {
	expires := time.Now().Add(time.Minute * 5)
	b, _ := app.EncryptId(id)
	cookie := http.Cookie{Name: "oauthstate", Value: b, Expires: expires}
	http.SetCookie(w, &cookie)

	return b
}
