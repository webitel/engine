package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitMember() {
	api.Router.Handle("cc_member_direct", api.ApiWebSocketHandler(api.memberDirect))
	api.Router.Handle("cc_fetch_offline_members", api.ApiWebSocketHandler(api.offlineMembers))
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
