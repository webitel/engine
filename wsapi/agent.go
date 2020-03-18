package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitAgent() {
	api.Router.Handle("cc_agent_session", api.ApiWebSocketHandler(api.getAgentSession))
	api.Router.Handle("cc_agent_subscribe_status", api.ApiWebSocketHandler(api.subscribeAgentsStatus))
	api.Router.Handle("cc_agent_waiting", api.ApiWebSocketHandler(api.loginAgent)) //FIXME /cc.AgentService/Login
	api.Router.Handle("cc_agent_logout", api.ApiWebSocketHandler(api.logoutAgent)) //FIXME /cc.AgentService/Login
	api.Router.Handle("cc_agent_pause", api.ApiWebSocketHandler(api.pauseAgent))   //FIXME /cc.AgentService/Login
}

func (api *API) subscribeAgentsStatus(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var ok bool

	h, e := api.App.GetHubById(req.Session.Domain(0)) //FIXME
	if e != nil {
		return nil, e
	}

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	return nil, h.SubscribeSessionAgentStatus(conn, int(agentId))
}

func (api *API) getAgentSession(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var userId int64
	var domainId int64
	var ok bool
	if userId, ok = req.Data["user_id"].(int64); !ok {
		userId = conn.UserId
	}

	if domainId, ok = req.Data["domain_id"].(int64); !ok {
		domainId = conn.DomainId
	}

	sess, err := api.ctrl.GetAgentSession(conn.GetSession(), domainId, userId)
	if err != nil {
		return nil, err
	}

	return sess.ToMap(), nil
}

func (api *API) loginAgent(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	err := api.ctrl.LoginAgent(conn.GetSession(), int64(domainId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) logoutAgent(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	err := api.ctrl.LogoutAgent(conn.GetSession(), int64(domainId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) pauseAgent(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var payload string
	var timeout float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	payload, _ = req.Data["payload"].(string)
	timeout, _ = req.Data["timeout"].(float64)

	err := api.ctrl.PauseAgent(conn.GetSession(), int64(domainId), int64(agentId), payload, int(timeout))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}
