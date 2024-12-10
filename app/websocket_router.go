package app

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

type webSocketHandler interface {
	ServeWebSocket(*WebConn, *model.WebSocketRequest)
}

type WebSocketRouter struct {
	app      *App
	handlers map[string]webSocketHandler
}

func (wr *WebSocketRouter) Handle(action string, handler webSocketHandler) {
	wr.handlers[action] = handler
}

func (wr *WebSocketRouter) ServeWebSocket(conn *WebConn, r *model.WebSocketRequest) {
	if r.Action == "" {
		err := model.NewBadRequestError("api.web_socket_router.no_action.app_error", "")
		ReturnWebSocketError(conn, r, err)
		return
	}

	if r.Seq <= 0 {
		err := model.NewBadRequestError("api.web_socket_router.bad_seq.app_error", "")
		ReturnWebSocketError(conn, r, err)
		return
	}

	if r.Action == model.WEBSOCKET_AUTHENTICATION_CHALLENGE {
		if conn.GetSessionToken() != "" {
			return
		}

		token, ok := r.Data["token"].(string)
		if !ok {
			conn.log.Error("not found token")
			conn.WebSocket.Close()
			return
		}

		conn.log.Debug("search session from token")

		session, err := wr.app.GetSession(token)
		if err != nil {
			ReturnWebSocketError(conn, r, err)
			return
		}

		conn.log.Debug("found session from token")

		if session.CountLicenses() == 0 {
			ReturnWebSocketError(conn, r, model.SocketPermissionError)
			return
		}

		conn.SetSession(session)
		conn.SetSessionToken(session.Token)
		conn.UserId = session.UserId
		conn.DomainId = session.DomainId

		wr.app.HubRegister(conn)

		resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, nil)
		conn.Send <- resp

		return
	}

	if !conn.IsAuthenticated() {
		err := model.NewInternalError("api.web_socket_router.not_authenticated.app_error", "")
		ReturnWebSocketError(conn, r, err)
		return
	}

	handler, ok := wr.handlers[r.Action]
	if !ok {
		err := model.NewBadRequestError("api.web_socket_router.bad_action.app_error", "")
		ReturnWebSocketError(conn, r, err)
		return
	}
	//FIXME
	go handler.ServeWebSocket(conn, r)
}

func ReturnWebSocketError(conn *WebConn, r *model.WebSocketRequest, err model.AppError) {
	conn.log.Error(err.Error(),
		wlog.Err(err),
	)

	errorResp := model.NewWebSocketError(r.Seq, err)
	conn.Send <- errorResp
}
