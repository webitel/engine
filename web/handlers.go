package web

import (
	"fmt"
	"net/http"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
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
	wlog.Debug(fmt.Sprintf("[%s] %v %v (%v)", ip, r.Method, r.URL.Path, r.Header.Get("User-Agent")))

	c := &Context{}
	c.App = h.App
	c.RequestId = model.NewId()
	c.UserAgent = r.UserAgent()
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
		c.Err.SetRequestId(c.RequestId)

		if c.Err.GetId() == "api.context.session_expired.app_error" {
			c.LogInfo(c.Err)
		} else {
			c.LogError(c.Err)
		}

		w.WriteHeader(c.Err.GetStatusCode())
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

func ReadUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

func ReadApplicationName(r *http.Request) string {
	return r.URL.Query().Get("application_name")
}

func ReadApplicationVersion(r *http.Request) string {
	return r.URL.Query().Get("ver")
}
