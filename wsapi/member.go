package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitMember() {
	api.Router.Handle("cc_member_direct", api.ApiWebSocketHandler(api.memberDirect))
}

func (api *API) memberDirect(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var agentId float64
	var domainId float64
	var memberId float64
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

	attemptId, err := api.ctrl.DirectAgentToMember(conn.GetSession(), int64(domainId), int64(memberId), int64(agentId))
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["attempt_id"] = attemptId
	return res, nil
}
