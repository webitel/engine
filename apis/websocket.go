package apis

import (
	"github.com/gorilla/websocket"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/web"
	"github.com/webitel/wlog"
	"net/http"
)

func (api *API) InitWebSocket() {
	api.Routes.Root.Handle("/websocket", api.ApiHandlerTrustRequester(connectWebSocket)).Methods("GET")
}

func connectWebSocket(c *Context, w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  model.SocketBufferSize,
		WriteBufferSize: model.SocketBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}

	log := c.App.Log.With(wlog.Namespace("context")).
		With(wlog.String("protocol", "wss"))

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	defer func() {
		c.App.Count.Add(-1)
		log.Debug("close socket", wlog.Int64("count", c.App.Count.Load()))
	}()

	c.App.Count.Add(1)
	log.Debug("open socket", wlog.Int64("count", c.App.Count.Load()))

	wc := c.App.NewWebConn(ws, c.Session, web.ReadApplicationName(r), web.ReadApplicationVersion(r),
		web.ReadUserAgent(r), web.ReadUserIP(r))
	wc.Pump()
}
