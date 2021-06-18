package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (api *API) InitMember() {
	api.Router.Handle("cc_member_direct", api.ApiWebSocketHandler(api.memberDirect))
	api.Router.Handle("cc_member_page", api.ApiWebSocketHandler(api.getMember))
	api.Router.Handle("cc_fetch_offline_members", api.ApiWebSocketHandler(api.offlineMembers))
	api.Router.Handle("cc_reporting", api.ApiWebSocketHandler(api.reporting))
	api.Router.Handle("cc_renewal", api.ApiWebSocketHandler(api.renewalAttempt))
}

func (api *API) renewalAttempt(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var attemptId, renewal float64
	var ok bool

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	if renewal, ok = req.Data["renewal_sec"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "renewal_sec")
	}

	if err := api.ctrl.RenewalAttempt(conn.GetSession(), int64(attemptId), uint32(renewal)); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) reporting(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var attemptId, agentId float64
	var ok bool
	var nextDistributeAt *int64
	var expire *int64
	var status string

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	if ok, _ = req.Data["success"].(bool); ok {
		status = "success" // TODO enum
	}

	description, _ := req.Data["description"].(string)
	display, _ := req.Data["display"].(bool)

	if tmp, ok := req.Data["next_distribute_at"].(float64); ok {
		nextDistributeAt = model.NewInt64(int64(tmp))
	}

	if tmp, ok := req.Data["nextDistributeAt"].(float64); nextDistributeAt != nil && ok {
		nextDistributeAt = model.NewInt64(int64(tmp))
	}

	if tmp, ok := req.Data["expire"].(float64); ok {
		expire = model.NewInt64(int64(tmp))
	}

	agentId, _ = req.Data["agent_id"].(float64)

	err := api.ctrl.ReportingAttempt(conn.GetSession(), int64(attemptId), status, description, nextDistributeAt, expire, nil, display, int32(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) memberDirect(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var memberId float64
	var communicationId float64
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	if memberId, ok = req.Data["member_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "member_id")
	}

	if communicationId, ok = req.Data["communication_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "communication_id")
	}

	attemptId, err := api.ctrl.DirectAgentToMember(conn.GetSession(), int64(domainId), int64(memberId), int(communicationId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["attempt_id"] = attemptId
	return res, nil
}

func (api *API) offlineMembers(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var page float64
	var perPage float64
	var q string
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	page, _ = req.Data["page"].(float64)
	perPage, _ = req.Data["per_page"].(float64)
	q, _ = req.Data["q"].(string)

	list, end, err := api.ctrl.ListOfflineQueueForAgent(conn.GetSession(), &model.SearchOfflineQueueMembers{
		ListRequest: model.ListRequest{
			Q:        q,
			Page:     int(page),
			PerPage:  int(perPage),
			DomainId: int64(domainId),
		},
		AgentId: int(agentId),
	})

	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["items"] = list
	res["next"] = !end
	return res, nil
}

func (api *API) getMember(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var queueId float64
	var memberId float64
	var ok bool
	var err *model.AppError
	session := conn.GetSession()

	if memberId, ok = req.Data["member_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "member_id")
	}

	if queueId, ok = req.Data["queue_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "queue_id")
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.App.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.App.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.App.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.Member
	out, err = api.App.GetMember(session.Domain(0), int64(queueId), int64(memberId))

	return model.InterfaceToMapString(out), nil
}
