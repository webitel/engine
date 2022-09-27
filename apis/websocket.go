package apis

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/webitel/engine/model"
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
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wlog.Error(fmt.Sprintf("websocket connect err: %v", err))
		return
	}

	defer func() {
		c.App.Count.Add(-1)
		wlog.Info(fmt.Sprintf("count socket %d", c.App.Count.Load()))
	}()

	c.App.Count.Add(1)
	wlog.Info(fmt.Sprintf("count socket %d", c.App.Count.Load()))

	wc := c.App.NewWebConn(ws, c.Session, c.T, "")
	wc.Pump()
}
