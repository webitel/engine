package call_manager

import (
	"fmt"
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
	return "sip:192.168.177.10"
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

func (cm *callManager) Bridge(legA, legANode, legB, legBNode string) {
	var cli CallClient
	var err *model.AppError

	cli, err = cm.getClient(legANode)
	if err != nil {
		panic(1)
	}

	if legANode == legBNode {
		_, err = cli.BridgeCall(legA, legB, "")
	} else {
		newUuid := model.NewUuid()

		r1 := &model.CallRequest{
			Endpoints: []string{"sofia/sip/bridge@test.com"},
			Variables: map[string]string{
				"sip_h_X-Webitel-ParentId": legB,
				"sip_route_uri":            "sip:10.10.10.25:5080",
				"origination_uuid":         newUuid,
				"ignore_early_media":       "true",
			},
			Timeout:      10,
			CallerName:   "",
			CallerNumber: "",
			Applications: []*model.CallRequestApplication{
				{
					AppName: "answer",
				},
				{
					AppName: "set",
					Args:    fmt.Sprintf("res=${uuid_bridge %s %s}", newUuid, legA),
				},
			},
		}

		_, _, err = cli.NewCall(r1)
	}

	if err != nil {
		fmt.Println("ERROR ", err.Error())
	} else {
		fmt.Println("OK")
	}
}
