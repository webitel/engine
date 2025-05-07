package client

import (
	"context"
	"github.com/webitel/engine/pkg/wbt"
	"github.com/webitel/engine/pkg/wbt/gen/cc"
	"github.com/webitel/wlog"
	"sync"
)

const ServiceName = "call_center"

type AgentApi interface {
	Online(domainId, agentId int64, onDemand bool) error
	Offline(domainId, agentId int64) error
	Pause(domainId, agentId int64, payload string, timeout int) error

	WaitingChannel(agentId int, channel string) (int64, error)

	AcceptTask(appId string, domainId, attemptId int64) error
	CloseTask(appId string, domainId, attemptId int64) error
	RunTrigger(ctx context.Context, domainId int64, userId int64, triggerId int32, vars map[string]string) (string, error)
}

type MemberApi interface {
	AttemptResult(result *cc.AttemptResultRequest) error
	RenewalResult(domainId, attemptId int64, renewal uint32) error

	JoinCallToQueue(ctx context.Context, in *cc.CallJoinToQueueRequest) (cc.MemberService_CallJoinToQueueClient, error)
	JoinChatToQueue(ctx context.Context, in *cc.ChatJoinToQueueRequest) (cc.MemberService_ChatJoinToQueueClient, error)
	CallJoinToAgent(ctx context.Context, in *cc.CallJoinToAgentRequest) (cc.MemberService_CallJoinToAgentClient, error)
	TaskJoinToAgent(ctx context.Context, in *cc.TaskJoinToAgentRequest) (cc.MemberService_TaskJoinToAgentClient, error)

	DirectAgentToMember(domainId int64, memberId int64, communicationId int, agentId int64) (int64, error)
	CancelAgentDistribute(ctx context.Context, in *cc.CancelAgentDistributeRequest) (*cc.CancelAgentDistributeResponse, error)
	ProcessingActionForm(ctx context.Context, in *cc.ProcessingFormActionRequest) (*cc.ProcessingFormActionResponse, error)
	ProcessingActionComponent(ctx context.Context, in *cc.ProcessingComponentActionRequest) (*cc.ProcessingComponentActionResponse, error)
	SaveFormFields(domainId, attemptId int64, fields map[string]string, form []byte) error
	CancelAttempt(ctx context.Context, attemptId int64, result, appId string) error
	InterceptAttempt(ctx context.Context, domainId int64, attemptId int64, agentId int32) error
	ResumeAttempt(ctx context.Context, attemptId int64, domainId int64) error
}

type CCManager interface {
	Start() error
	Stop()

	Agent() AgentApi
	Member() MemberApi
}

type ccManager struct {
	startOnce  sync.Once
	consulAddr string

	agentClient  *wbt.Client[cc.AgentServiceClient]
	memberClient *wbt.Client[cc.MemberServiceClient]

	agent  AgentApi
	member MemberApi
}

func NewCCManager(consulAddr string) CCManager {
	cli := &ccManager{
		consulAddr: consulAddr,
	}

	return cli
}

func (cm *ccManager) Agent() AgentApi {
	return cm.agent
}

func (cm *ccManager) Member() MemberApi {
	return cm.member
}

func (cm *ccManager) Start() error {
	wlog.Debug("starting cc service")
	var err error

	cm.startOnce.Do(func() {
		cm.agentClient, err = wbt.NewClient(cm.consulAddr, ServiceName, cc.NewAgentServiceClient)
		if err != nil {
			return
		}

		cm.memberClient, err = wbt.NewClient(cm.consulAddr, ServiceName, cc.NewMemberServiceClient)
		if err != nil {
			return
		}

		cm.agent = NewAgentApi(cm.agentClient)
		cm.member = NewMemberApi(cm.memberClient)
	})
	return err
}

func (cm *ccManager) Stop() {

}
