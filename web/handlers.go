package web

import (
	"fmt"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"net/http"
)

type Handler struct {
	App            *app.App
	HandleFunc     func(*Context, http.ResponseWriter, *http.Request)
	RequireSession bool
	TrustRequester bool
	RequireMfa     bool
	IsStatic       bool
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := ReadUserIP(r)
	wlog.Debug(fmt.Sprintf("[%s] %v %v", ip, r.Method, r.URL.Path))

	c := &Context{}
	c.App = h.App
	c.RequestId = model.NewId()
	c.UserAgent = r.UserAgent()
	c.T, _ = localization.GetTranslationsAndLocale(w, r)
	c.AcceptLanguage = r.Header.Get("Accept-Language")
	c.IpAddress = ip
	c.Log = h.App.Log

	w.Header().Set(model.HEADER_REQUEST_ID, c.RequestId)
	w.Header().Set("Content-Type", "application/json")

	if c.Err == nil && h.RequireSession {
		c.SessionRequired()
	}

	//
	if c.Err == nil {
		h.HandleFunc(c, w, r)
	}

	if c.Err != nil {
		c.Err.Translate(c.T)
		c.Err.RequestId = c.RequestId

		if c.Err.Id == "api.context.session_expired.app_error" {
			c.LogInfo(c.Err)
		} else {
			c.LogError(c.Err)
		}

		c.Err.Where = r.URL.Path

		w.WriteHeader(c.Err.StatusCode)
		w.Write([]byte(c.Err.ToJson()))
	}
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
