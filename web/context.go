package web

import (
	"net/http"

	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/wlog"
)

type Context struct {
	App            *app.App
	Log            *wlog.Logger
	Session        auth_manager.Session
	RequestId      string
	IpAddress      string
	UserAgent      string
	AcceptLanguage string
	T              i18n.TranslateFunc
	Err            model.AppError
}

func (c *Context) LogError(err model.AppError) {
	// Filter out 404s, endless reconnects and browser compatibility errors
	if err.GetStatusCode() == http.StatusNotFound {
		c.LogDebug(err)
	} else {
		c.Log.Error(
			err.SystemMessage(localization.TDefault),
			wlog.Int("http_code", err.GetStatusCode()),
			wlog.String("err_details", err.GetDetailedError()),
		)
	}
}

func (c *Context) LogInfo(err model.AppError) {
	// Filter out 401s
	if err.GetStatusCode() == http.StatusUnauthorized {
		c.LogDebug(err)
	} else {
		c.Log.Info(
			err.SystemMessage(localization.TDefault),
			wlog.Int("http_code", err.GetStatusCode()),
			wlog.String("err_details", err.GetDetailedError()),
		)
	}
}

func (c *Context) LogDebug(err model.AppError) {
	c.Log.Debug(
		err.SystemMessage(localization.TDefault),
		wlog.Int("http_code", err.GetStatusCode()),
		wlog.String("err_details", err.GetDetailedError()),
	)
}

func (c *Context) SessionRequired() {
	if len(c.Session.Id) == 0 {
		c.Err = model.NewInternalError("api.context.session_expired.app_error", "UserRequired")
		return
	}
}
