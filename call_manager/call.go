package call_manager

import (
	"github.com/webitel/call_center/discovery"
	"github.com/webitel/engine/model"
	"net/http"
)

type Call interface {
}

func (cm *callManager) getClient(nodeId string) (CallClient, *model.AppError) {
	var err error
	var conn discovery.Connection
	if nodeId != "" {
		conn, err = cm.poolConnections.GetById(nodeId)
	} else {
		conn, err = cm.poolConnections.Get(discovery.StrategyRoundRobin)
	}
	if err != nil {
		return nil, model.NewAppError("CallManager", "call_manager.get_cli.by_id.app_err", nil, err.Error(), http.StatusInternalServerError)
	}
	return conn.(CallClient), nil

}

//FIXME
func (cm *callManager) SipRouteUri() string {
	return "sip:192.168.177.9"
}

func (cm *callManager) MakeOutboundCall(req *model.CallRequest) (string, *model.AppError) {
	cli, err := cm.getClient("")
	if err != nil {
		return "", err
	}
	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	req.Variables["sip_route_uri"] = cm.SipRouteUri()

	uuid, cause, err := cli.NewCall(req)
	if err != nil {
		return "", err
	}
	if cause != "" {
		//FIXME
	}

	return uuid, nil
}
