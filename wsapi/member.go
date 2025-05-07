package wsapi

import (
	"context"
	"encoding/json"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (api *API) InitMember() {
	api.Router.Handle("cc_member_direct", api.ApiWebSocketHandler(api.memberDirect))
	api.Router.Handle("cc_member_page", api.ApiWebSocketHandler(api.getMember))
	api.Router.Handle("cc_fetch_offline_members", api.ApiWebSocketHandler(api.offlineMembers))
	api.Router.Handle("cc_reporting", api.ApiWebSocketHandler(api.reporting))
	api.Router.Handle("cc_renewal", api.ApiWebSocketHandler(api.renewalAttempt))
	api.Router.Handle("cc_form_action", api.ApiWebSocketHandler(api.processingActionFormAttempt))
	api.Router.Handle("cc_component_action", api.ApiWebSocketHandler(api.processingActionComponentAttempt))
	api.Router.Handle("cc_form_save", api.ApiWebSocketHandler(api.processingSaveFormAttempt))
	api.Router.Handle("cc_intercept_attempt", api.ApiWebSocketHandler(api.interceptAttempt))
}

func (api *API) interceptAttempt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var attemptId, agentId float64
	var ok bool

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if err := api.ctrl.InterceptAttempt(conn.GetSession(), int64(attemptId), int32(agentId)); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) renewalAttempt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
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

func (api *API) processingActionFormAttempt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var attemptId float64
	var ok bool
	var appId, formId, action string
	var fields map[string]interface{}

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	if appId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	action, _ = req.Data["action"].(string)
	fields, _ = req.Data["fields"].(map[string]interface{})

	if err := api.ctrl.ProcessingActionFormAttempt(conn.GetSession(), int64(attemptId), appId,
		formId, action, model.MapInterfaceToString(fields)); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) processingActionComponentAttempt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var attemptId float64
	var ok, sync bool
	var appId, formId, component, action string
	var vars map[string]interface{}

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	if appId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	action, _ = req.Data["action"].(string)
	component, _ = req.Data["componentId"].(string)
	formId, _ = req.Data["formId"].(string)
	sync, _ = req.Data["sync"].(bool)
	vars, _ = req.Data["vars"].(map[string]interface{})

	if err := api.ctrl.ProcessingActionComponentAttempt(conn.GetSession(), int64(attemptId), appId, formId, component,
		action, model.MapInterfaceToString(vars), sync); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) processingSaveFormAttempt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var attemptId float64
	var ok bool
	var fields map[string]interface{}
	var form []byte

	if attemptId, ok = req.Data["attempt_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	// TODO app_id for node
	if _, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	fields, _ = req.Data["fields"].(map[string]interface{})
	if v, ok := req.Data["form"]; ok {
		form, _ = json.Marshal(v)
	}

	if err := api.ctrl.ProcessingSaveForm(conn.GetSession(), int64(attemptId), model.MapInterfaceToString(fields), form); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) reporting(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var attemptId, agentId float64
	var ok bool
	var nextDistributeAt *int64
	var expire *int64
	var waitBetweenRetries *int32
	var status string
	var exclDes, onlyComm bool

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

	if tmp, ok := req.Data["nextDistributeAt"].(float64); nextDistributeAt == nil && ok {
		nextDistributeAt = model.NewInt64(int64(tmp))
	}

	if tmp, ok := req.Data["expire"].(float64); ok {
		expire = model.NewInt64(int64(tmp))
	}

	if tmp, ok := req.Data["waitBetweenRetries"].(float64); ok {
		waitBetweenRetries = model.NewInt32(int32(tmp))
	}

	exclDes, _ = req.Data["exclude_current_communication"].(bool)
	onlyComm, _ = req.Data["only_current_communication"].(bool)

	agentId, _ = req.Data["agent_id"].(float64)

	err := api.ctrl.ReportingAttempt(conn.GetSession(), int64(attemptId), status, description, nextDistributeAt, expire,
		nil, display, int32(agentId), exclDes, waitBetweenRetries, onlyComm)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) memberDirect(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
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

	attemptId, err := api.ctrl.DirectAgentToMember(ctx, conn.GetSession(), int64(domainId),
		int64(memberId), int(communicationId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["attempt_id"] = attemptId
	return res, nil
}

func (api *API) offlineMembers(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var agentId float64
	var domainId float64
	var page float64
	var size float64
	var q string
	var ok bool

	if agentId, ok = req.Data["agent_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "agent_id")
	}

	if domainId, ok = req.Data["domain_id"].(float64); !ok {
		domainId = float64(conn.DomainId)
	}

	page, _ = req.Data["page"].(float64)
	size, _ = req.Data["size"].(float64)
	q, _ = req.Data["q"].(string)

	list, end, err := api.ctrl.ListOfflineQueueForAgent(ctx, conn.GetSession(), &model.SearchOfflineQueueMembers{
		ListRequest: model.ListRequest{
			Q:        q,
			Page:     int(page),
			PerPage:  int(size),
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

func (api *API) getMember(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var queueId float64
	var memberId float64
	var ok bool
	var err model.AppError
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
		if perm, err = api.App.QueueCheckAccess(conn.Ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.App.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.Member
	out, err = api.App.GetMember(ctx, session.Domain(0), int64(queueId), int64(memberId))

	return model.InterfaceToMapString(out), nil
}
