package cc

import (
	"context"
	"github.com/webitel/engine/pkg/wbt"
	"github.com/webitel/engine/pkg/wbt/gen/cc"
)

type agentApi struct {
	*wbt.Client[cc.AgentServiceClient]
}

func NewAgentApi(c *wbt.Client[cc.AgentServiceClient]) AgentApi {
	return &agentApi{
		Client: c,
	}
}

func (api *agentApi) Online(domainId, agentId int64, onDemand bool) error {
	_, err := api.Api.Online(context.TODO(), &cc.OnlineRequest{
		AgentId:  agentId,
		OnDemand: onDemand,
		DomainId: domainId,
	})
	return err
}

func (api *agentApi) Offline(domainId, agentId int64) error {
	_, err := api.Api.Offline(context.TODO(), &cc.OfflineRequest{
		AgentId:  agentId,
		DomainId: domainId,
	})
	return err
}

func (api *agentApi) Pause(domainId, agentId int64, payload string, timeout int) error {

	_, err := api.Api.Pause(context.TODO(), &cc.PauseRequest{
		AgentId:  agentId,
		Payload:  payload,
		Timeout:  int32(timeout),
		DomainId: domainId,
	})
	return err
}

func (api *agentApi) WaitingChannel(agentId int, channel string) (int64, error) {
	if res, err := api.Api.WaitingChannel(context.TODO(), &cc.WaitingChannelRequest{
		AgentId: int32(agentId),
		Channel: channel,
	}); err != nil {
		return 0, err
	} else {
		return res.Timestamp, nil
	}

}

func (api *agentApi) AcceptTask(appId string, domainId, attemptId int64) error {
	ctx := api.StaticHost(context.Background(), appId)

	_, err := api.Api.AcceptTask(ctx, &cc.AcceptTaskRequest{
		Id:       attemptId,
		AppId:    appId,
		DomainId: domainId,
	})

	return err
}

func (api *agentApi) CloseTask(appId string, domainId, attemptId int64) error {
	ctx := api.StaticHost(context.Background(), appId)

	_, err := api.Api.CloseTask(ctx, &cc.CloseTaskRequest{
		Id:       attemptId,
		AppId:    appId,
		DomainId: domainId,
	})

	return err
}

func (api *agentApi) RunTrigger(ctx context.Context, domainId int64, userId int64, triggerId int32, vars map[string]string) (string, error) {
	res, err := api.Api.RunTrigger(ctx, &cc.RunTriggerRequest{
		DomainId:  domainId,
		TriggerId: triggerId,
		UserId:    userId,
		Variables: vars,
	})

	if err != nil {
		return "", err
	}

	return res.JobId, nil
}
