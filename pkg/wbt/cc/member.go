package client

import (
	"context"
	"github.com/webitel/engine/pkg/wbt"
	"github.com/webitel/engine/pkg/wbt/gen/cc"
)

type memberApi struct {
	*wbt.Client[cc.MemberServiceClient]
}

func NewMemberApi(c *wbt.Client[cc.MemberServiceClient]) MemberApi {
	return &memberApi{
		Client: c,
	}
}

func (api *memberApi) JoinCallToQueue(ctx context.Context, in *cc.CallJoinToQueueRequest) (cc.MemberService_CallJoinToQueueClient, error) {
	return api.Api.CallJoinToQueue(ctx, in)
}

func (api *memberApi) JoinChatToQueue(ctx context.Context, in *cc.ChatJoinToQueueRequest) (cc.MemberService_ChatJoinToQueueClient, error) {
	return api.Api.ChatJoinToQueue(ctx, in)
}

func (api *memberApi) DirectAgentToMember(domainId int64, memberId int64, communicationId int, agentId int64) (int64, error) {

	res, err := api.Api.DirectAgentToMember(context.Background(), &cc.DirectAgentToMemberRequest{
		MemberId:        memberId,
		AgentId:         agentId,
		CommunicationId: int32(communicationId),
		DomainId:        domainId,
	})

	if err != nil {
		return 0, err
	}

	return res.AttemptId, nil
}

func (api *memberApi) AttemptResult(result *cc.AttemptResultRequest) error {

	_, err := api.Api.AttemptResult(context.Background(), result)

	if err != nil {
		return err
	}

	return nil
}

func (api *memberApi) RenewalResult(domainId, attemptId int64, renewal uint32) error {
	_, err := api.Api.AttemptRenewalResult(context.Background(), &cc.AttemptRenewalResultRequest{
		DomainId:  domainId,
		AttemptId: attemptId,
		Renewal:   renewal,
	})

	return err
}

func (api *memberApi) CallJoinToAgent(ctx context.Context, in *cc.CallJoinToAgentRequest) (cc.MemberService_CallJoinToAgentClient, error) {
	return api.Api.CallJoinToAgent(ctx, in)
}

func (api *memberApi) TaskJoinToAgent(ctx context.Context, in *cc.TaskJoinToAgentRequest) (cc.MemberService_TaskJoinToAgentClient, error) {
	return api.Api.TaskJoinToAgent(ctx, in)
}

func (api *memberApi) CancelAgentDistribute(ctx context.Context, in *cc.CancelAgentDistributeRequest) (*cc.CancelAgentDistributeResponse, error) {
	return api.Api.CancelAgentDistribute(ctx, in)
}

func (api *memberApi) ProcessingActionForm(ctx context.Context, in *cc.ProcessingFormActionRequest) (*cc.ProcessingFormActionResponse, error) {
	ctx2 := api.StaticHost(ctx, in.AppId)
	return api.Api.ProcessingFormAction(ctx2, in)
}

func (api *memberApi) ProcessingActionComponent(ctx context.Context, in *cc.ProcessingComponentActionRequest) (*cc.ProcessingComponentActionResponse, error) {
	ctx2 := api.StaticHost(ctx, in.AppId)
	return api.Api.ProcessingComponentAction(ctx2, in)
}

func (api *memberApi) CancelAttempt(ctx context.Context, attemptId int64, result, appId string) error {
	ctx2 := api.StaticHost(ctx, appId)

	_, err := api.Api.CancelAttempt(ctx2, &cc.CancelAttemptRequest{
		AttemptId: attemptId,
		Result:    result,
	})

	return err
}

func (api *memberApi) InterceptAttempt(ctx context.Context, domainId int64, attemptId int64, agentId int32) error {
	_, err := api.Api.InterceptAttempt(ctx, &cc.InterceptAttemptRequest{
		DomainId:  domainId,
		AttemptId: attemptId,
		AgentId:   agentId,
	})

	return err
}

func (api *memberApi) ResumeAttempt(ctx context.Context, attemptId int64, domainId int64) error {
	_, err := api.Api.ResumeAttempt(ctx, &cc.ResumeAttemptRequest{
		DomainId:  domainId,
		AttemptId: attemptId,
	})

	return err
}

func (api *memberApi) SaveFormFields(domainId, attemptId int64, fields map[string]string, form []byte) error {
	_, err := api.Api.ProcessingFormSave(context.Background(), &cc.ProcessingFormSaveRequest{
		DomainId:  domainId,
		AttemptId: attemptId,
		Fields:    fields,
		Form:      form,
	})

	return err
}
