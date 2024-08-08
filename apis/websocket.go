package apis

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/webitel/webitel-go-kit/logging/wlog"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/web"
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
		EnableCompression: true,
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

	wc := c.App.NewWebConn(ws, c.Session, c.T, "", web.ReadUserIP(r))
	wc.Pump()
}
