package wsapi

import (
	"context"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitAgent() {
	api.Router.Handle("cc_agent_session", api.ApiWebSocketHandler(api.getAgentSession))
	api.Router.Handle("cc_agent_subscribe_status", api.ApiWebSocketHandler(api.subscribeAgentsStatus))
	api.Router.Handle("cc_agent_online", api.ApiWebSocketHandler(api.onlineAgent))
	api.Router.Handle("cc_agent_waiting", api.ApiWebSocketHandler(api.waitingAgent))
	api.Router.Handle("cc_agent_offline", api.ApiWebSocketHandler(api.offlineAgent))
	api.Router.Handle("cc_agent_pause", api.ApiWebSocketHandler(api.pauseAgent))
	api.Router.Handle("cc_agent_tasks", api.ApiWebSocketHandler(api.agentTasks))
	api.Router.Handle("cc_agent_task_accept", api.ApiWebSocketHandler(api.acceptAgentTask))
	api.Router.Handle("cc_agent_task_close", api.ApiWebSocketHandler(api.closeAgentTask))
}

func (api *API) subscribeAgentsStatus(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
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

func (api *API) getAgentSession(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var userId int64
	var domainId int64
	var ok bool
	if userId, ok = req.Data["user_id"].(int64); !ok {
		userId = conn.UserId
	}

	if domainId, ok = req.Data["domain_id"].(int64); !ok {
		domainId = conn.DomainId
	}

	sess, err := api.ctrl.GetAgentSession(ctx, conn.GetSession(), domainId, userId)
	if err != nil {
		return nil, err
	}

	return sess.ToMap(), nil
}

func (api *API) onlineAgent(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId float64
	var domainId float64
	var onDemand bool
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	onDemand, _ = req.Data["on_demand"].(bool)
	err := api.ctrl.LoginAgent(ctx, conn.GetSession(), int64(domainId), int64(agentId), onDemand)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) offlineAgent(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId float64
	var domainId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	err := api.ctrl.LogoutAgent(ctx, conn.GetSession(), int64(domainId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) pauseAgent(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]any, model.AppError) {
	var agentId float64
	var domainId float64
	var payload, statusComment string
	var timeout float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	payload, _ = req.Data["payload"].(string)
	statusComment, _ = req.Data["status_comment"].(string)
	timeout, _ = req.Data["timeout"].(float64)

	err := api.ctrl.PauseAgent(ctx, conn.GetSession(), int64(domainId), int64(agentId), payload, statusComment, int(timeout))
	if err != nil {
		return nil, err
	}

	res := make(map[string]any)
	return res, nil
}

func (api *API) waitingAgent(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId float64
	var domainId float64
	var channel string
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if channel, ok = req.Data["channel"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	timestamp, err := api.ctrl.WaitingAgent(ctx, conn.GetSession(), int64(domainId), int64(agentId), channel)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["timestamp"] = timestamp
	return res, nil
}

func (api *API) agentTasks(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId, domainId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	list, err := api.ctrl.ActiveAgentTasks(ctx, conn.GetSession(), int64(domainId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["items"] = list
	return res, nil
}

func (api *API) acceptAgentTask(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId, attemptId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	appId, _ := req.Data["app_id"].(string)
	if appId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if agentId == 0 {
		//todo
	}

	err := api.ctrl.AcceptAgentTask(conn.GetSession(), appId, int64(attemptId))
	return nil, err
}

func (api *API) closeAgentTask(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId, attemptId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	appId, _ := req.Data["app_id"].(string)
	if appId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if agentId == 0 {
		//todo
	}

	err := api.ctrl.CloseAgentTask(conn.GetSession(), appId, int64(attemptId))
	return nil, err
}
