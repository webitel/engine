package web

import (
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"net/http"
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
	Err            *model.AppError
}

func (c *Context) LogError(err *model.AppError) {
	// Filter out 404s, endless reconnects and browser compatibility errors
	if err.StatusCode == http.StatusNotFound {
		c.LogDebug(err)
	} else {
		c.Log.Error(
			err.SystemMessage(localization.TDefault),
			wlog.String("err_where", err.Where),
			wlog.Int("http_code", err.StatusCode),
			wlog.String("err_details", err.DetailedError),
		)
	}
}

func (c *Context) LogInfo(err *model.AppError) {
	// Filter out 401s
	if err.StatusCode == http.StatusUnauthorized {
		c.LogDebug(err)
	} else {
		c.Log.Info(
			err.SystemMessage(localization.TDefault),
			wlog.String("err_where", err.Where),
			wlog.Int("http_code", err.StatusCode),
			wlog.String("err_details", err.DetailedError),
		)
	}
}

func (c *Context) LogDebug(err *model.AppError) {
	c.Log.Debug(
		err.SystemMessage(localization.TDefault),
		wlog.String("err_where", err.Where),
		wlog.Int("http_code", err.StatusCode),
		wlog.String("err_details", err.DetailedError),
	)
}

func (c *Context) SessionRequired() {
	if len(c.Session.Id) == 0 {
		c.Err = model.NewAppError("", "api.context.session_expired.app_error", nil, "UserRequired", http.StatusUnauthorized)
		return
	}
}
