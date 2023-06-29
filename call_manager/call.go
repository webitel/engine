package call_manager

import (
	"encoding/json"
	"fmt"

	"github.com/webitel/engine/discovery"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

type Call interface {
}

func (cm *callManager) getClient(nodeId string) (CallClient, model.AppError) {
	var err error
	var conn discovery.Connection
	if nodeId != "" {
		conn, err = cm.poolConnections.GetById(nodeId)
	} else {
		conn, err = cm.poolConnections.Get(discovery.StrategyRoundRobin)
	}
	if err != nil {
		return nil, model.NewInternalError("call_manager.get_cli.by_id.app_err", err.Error())
	}
	return conn.(CallClient), nil

}

func (cm *callManager) MakeOutboundCall(req *model.CallRequest) (string, model.AppError) {
	cli, err := cm.getClient("")
	if err != nil {
		return "", err
	}
	return cli.MakeOutboundCall(req)
}

func DUMP(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	wlog.Error(string(s))
	return string(s)
}

func (cm *callManager) Bridge(legA, legANode, legB, legBNode string) model.AppError {
	var cli, cli2 CallClient
	var err model.AppError

	cli, err = cm.getClient(legANode)
	if err != nil {
		return err
	}

	if legANode == legBNode {
		_, err = cli.BridgeCall(legA, legB, "")
	} else {

		cli2, err = cm.getClient(legBNode)
		if err != nil {
			panic(1)
		}

		cli.SetCallVariables(legA, map[string]string{
			"sip_h_X-Webitel-Fwd": fmt.Sprintf("sip:w@%s:5080", cli2.Host()),
		})

		fmt.Println(fmt.Sprintf("%s sip:w@%s:5080", legA, cli2.Host()))
		err = cli.Execute("uuid_deflect", fmt.Sprintf("%s sip:w@%s:5080", legA, cli2.Host()))
	}

	return err
}
