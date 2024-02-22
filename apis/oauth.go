package apis

import (
	"fmt"
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

func (api *API) InitOAuth() {
	api.Routes.Endpoint.Handle("/oauth2/{id}/callback", api.ApiHandlerTrustRequester(handleOAuth2Callback)).Methods("GET")
}

func handleOAuth2Callback(c *Context, w http.ResponseWriter, r *http.Request) {
	var domainId, profileId int64
	var err model.AppError

	state := strings.Split(r.FormValue("state"), "::")
	if len(state) != 2 {
		// ERROR
	}

	domainId, err = c.App.DecryptId(state[0])
	if err != nil {
		c.Err = err
		return
	}

	profileId, err = c.App.DecryptId(state[1])
	if err != nil {
		c.Err = err
		return
	}

	p, err := c.App.Store.EmailProfile().Get(r.Context(), domainId, int(profileId))
	if err != nil {
		c.Err = err
		return
	}

	e, err := p.Oauth()
	if err != nil {
		c.Err = err
		return
	}
	code := r.FormValue("code")

	if code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(r.Form.Get("error_description")))
		return
	}

	token, err2 := e.Exchange(oauth2.NoContext, code)
	if err2 != nil {
		c.Err = model.NewBadRequestError("api.oauth2.callback.bad_request", err2.Error())
		return
	}

	if c.Err = c.App.EmailLoginOAuth(r.Context(), int(profileId), token); c.Err != nil {
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/integrations/email-profile/%d", *c.App.Config().PublicHostName, profileId), http.StatusSeeOther)
}
