package app

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/protos/cc"
)

func (app *App) CreateMember(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError) {
	q, err := app.GetQueueById(ctx, domainId, member.QueueId)
	if err != nil {
		return nil, err
	}
	if q.Type == 1 || q.Type == 6 {
		return nil, model.NewBadRequestError("app.member.valid.queue", "Mismatch queue type")
	}
	return app.Store.Member().Create(ctx, domainId, member)
}

func (a *App) SearchMembers(ctx context.Context, domainId int64, search *model.SearchMemberRequest) ([]*model.Member, bool, model.AppError) {
	list, err := a.Store.Member().SearchMembers(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) BulkCreateMember(ctx context.Context, domainId, queueId int64, fileName string, members []*model.Member) ([]int64, model.AppError) {
	q, err := app.GetQueueById(ctx, domainId, queueId)
	if err != nil {
		return nil, err
	}
	if q.Type == 1 || q.Type == 6 {
		return nil, model.NewBadRequestError("app.member.valid.queue", "Mismatch queue type")
	}

	if len(fileName) > 120 {
		return nil, model.NewBadRequestError("app.member.valid.file_name", "The filename can not be more than 120 symbols")
	}

	return app.Store.Member().BulkCreate(ctx, domainId, queueId, fileName, members)
}

func (app *App) GetMember(ctx context.Context, domainId, queueId, id int64) (*model.Member, model.AppError) {
	return app.Store.Member().Get(ctx, domainId, queueId, id)
}

func (app *App) UpdateMember(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError) {
	oldMember, err := app.GetMember(ctx, domainId, member.QueueId, member.Id)
	if err != nil {
		return nil, err
	}

	oldMember.Priority = member.Priority
	oldMember.ExpireAt = member.ExpireAt
	oldMember.Variables = member.Variables
	oldMember.Name = member.Name
	oldMember.Timezone = member.Timezone
	oldMember.Communications = member.Communications
	oldMember.Bucket = member.Bucket
	oldMember.MinOfferingAt = member.MinOfferingAt

	if oldMember.StopCause != nil && (member.StopCause == nil || *member.StopCause == "") {
		oldMember.ResetAttempts()
	}

	oldMember.StopCause = member.StopCause
	oldMember.Agent = member.Agent
	oldMember.Skill = member.Skill

	oldMember, err = app.Store.Member().Update(ctx, domainId, oldMember)
	if err != nil {
		return nil, err
	}

	return oldMember, nil
}

func (app *App) MaxMemberCommunications() int {
	return app.config.MaxMemberCommunications
}

func (app *App) PatchMember(ctx context.Context, domainId, queueId, id int64, patch *model.MemberPatch) (*model.Member, model.AppError) {
	oldMember, err := app.GetMember(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldMember.Patch(patch)

	if err = oldMember.IsValid(app.MaxMemberCommunications()); err != nil {
		return nil, err
	}

	oldMember, err = app.Store.Member().Update(ctx, domainId, oldMember)
	if err != nil {
		return nil, err
	}

	return oldMember, nil
}

func (app *App) RemoveMember(ctx context.Context, domainId, queueId, id int64) (*model.Member, model.AppError) {
	member, err := app.GetMember(ctx, domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.Member().Delete(ctx, queueId, id)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (app *App) RemoveMultiMembers(ctx context.Context, domainId int64, del *model.MultiDeleteMembers) ([]*model.Member, model.AppError) {
	return app.Store.Member().MultiDelete(ctx, del)
}

func (app *App) ResetMembers(ctx context.Context, domainId int64, req *model.ResetMembers) (int64, model.AppError) {
	return app.Store.Member().ResetMembers(ctx, domainId, req)
}

func (app *App) GetMemberAttempts(ctx context.Context, memberId int64) ([]*model.MemberAttempt, model.AppError) {
	return app.Store.Member().AttemptsList(ctx, memberId)
}

func (app *App) SearchAttemptsHistory(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, bool, model.AppError) {
	list, err := app.Store.Member().SearchAttemptsHistory(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) SearchAttempts(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.Attempt, bool, model.AppError) {
	list, err := app.Store.Member().SearchAttempts(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) DirectAgentToMember(domainId, memberId int64, communicationId int, agentId int64) (int64, model.AppError) {
	attemptId, err := app.cc.Member().DirectAgentToMember(domainId, memberId, communicationId, agentId)
	if err != nil {
		return 0, model.NewBadRequestError("app.cc_member.direct_agent.app_err", err.Error())
	}

	return attemptId, nil
}

func (app *App) ListOfflineQueueForAgent(ctx context.Context, domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, bool, model.AppError) {
	list, err := app.Store.Member().ListOfflineQueueForAgent(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) ReportingAttempt(attemptId int64, status, description string, nextOffering *int64, expireAt *int64, vars map[string]string,
	stickyDisplay bool, agentId int32, excludeDes bool) model.AppError {

	res := &cc.AttemptResultRequest{
		AttemptId:                   attemptId,
		Status:                      status,
		NextDistributeAt:            0,
		ExpireAt:                    0,
		Variables:                   vars,
		Display:                     stickyDisplay,
		Description:                 description,
		TransferQueueId:             0,
		AgentId:                     agentId,
		ExcludeCurrentCommunication: excludeDes,
	}

	if expireAt != nil {
		res.ExpireAt = *expireAt
	}

	if nextOffering != nil {
		res.NextDistributeAt = *nextOffering
	}

	err := app.cc.Member().AttemptResult(res)

	if err != nil {
		return model.NewBadRequestError("app.cc_member.reporting.app_err", err.Error())
	}

	return nil
}

func (app *App) RenewalAttempt(domainId, attemptId int64, renewal uint32) model.AppError {
	err := app.cc.Member().RenewalResult(domainId, attemptId, renewal)
	if err != nil {
		return model.NewBadRequestError("app.cc_member.renewal_attempt.app_err", err.Error())
	}

	return nil
}

func (app *App) ProcessingActionForm(domainId, attemptId int64, appId string, formId string, action string, fields map[string]string) model.AppError {
	_, err := app.cc.Member().ProcessingActionForm(context.Background(), &cc.ProcessingFormActionRequest{
		DomainId:  domainId,
		AttemptId: attemptId,
		AppId:     appId,
		FormId:    formId,
		Action:    action,
		Fields:    fields,
	})

	if err != nil {
		return model.NewBadRequestError("app.cc_member.form_action.app_err", err.Error())
	}

	return nil
}

func (app *App) InterceptAttempt(domainId, attemptId int64, agentId int32) model.AppError {
	err := app.cc.Member().InterceptAttempt(context.Background(), domainId, attemptId, agentId)
	if err != nil {
		return model.NewBadRequestError("app.cc_member.intercept.app_err", err.Error())
	}

	return nil
}
