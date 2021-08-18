package app

import (
	"github.com/webitel/engine/model"
	"net/http"
)

func (app *App) CreateMember(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	return app.Store.Member().Create(domainId, member)
}

func (a *App) SearchMembers(domainId int64, search *model.SearchMemberRequest) ([]*model.Member, bool, *model.AppError) {
	list, err := a.Store.Member().SearchMembers(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) BulkCreateMember(domainId, queueId int64, members []*model.Member) ([]int64, *model.AppError) {
	_, err := app.GetQueueById(domainId, queueId)
	if err != nil {
		return nil, err
	}
	return app.Store.Member().BulkCreate(domainId, queueId, members)
}

func (app *App) GetMemberPage(domainId, queueId int64, search *model.SearchMemberRequest) ([]*model.Member, bool, *model.AppError) {
	list, err := app.Store.Member().GetAllPage(domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetMember(domainId, queueId, id int64) (*model.Member, *model.AppError) {
	return app.Store.Member().Get(domainId, queueId, id)
}

func (app *App) UpdateMember(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	oldMember, err := app.GetMember(domainId, member.QueueId, member.Id)
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
	oldMember.StopCause = member.StopCause
	oldMember.Agent = member.Agent
	oldMember.Skill = member.Skill

	oldMember, err = app.Store.Member().Update(domainId, oldMember)
	if err != nil {
		return nil, err
	}

	return oldMember, nil
}

func (app *App) PatchMember(domainId, queueId, id int64, patch *model.MemberPatch) (*model.Member, *model.AppError) {
	oldMember, err := app.GetMember(domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldMember.Patch(patch)

	if err = oldMember.IsValid(); err != nil {
		return nil, err
	}

	oldMember, err = app.Store.Member().Update(domainId, oldMember)
	if err != nil {
		return nil, err
	}

	return oldMember, nil
}

func (app *App) RemoveMember(domainId, queueId, id int64) (*model.Member, *model.AppError) {
	member, err := app.GetMember(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.Member().Delete(queueId, id)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (app *App) RemoveMultiMembers(domainId, queueId int64, ids []int64, buckets []int64, cause []string, agentIds []int32) ([]*model.Member, *model.AppError) {
	return app.Store.Member().MultiDelete(queueId, ids, buckets, cause, agentIds)
}

func (app *App) GetMemberAttempts(memberId int64) ([]*model.MemberAttempt, *model.AppError) {
	return app.Store.Member().AttemptsList(memberId)
}

func (app *App) SearchAttemptsHistory(domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, bool, *model.AppError) {
	list, err := app.Store.Member().SearchAttemptsHistory(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) SearchAttempts(domainId int64, search *model.SearchAttempts) ([]*model.Attempt, bool, *model.AppError) {
	list, err := app.Store.Member().SearchAttempts(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) DirectAgentToMember(domainId, memberId int64, communicationId int, agentId int64) (int64, *model.AppError) {
	attemptId, err := app.cc.Member().DirectAgentToMember(domainId, memberId, communicationId, agentId)
	if err != nil {
		return 0, model.NewAppError("DirectAgentToMember", "app.cc_member.direct_agent.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return attemptId, nil
}

func (app *App) ListOfflineQueueForAgent(domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, bool, *model.AppError) {
	list, err := app.Store.Member().ListOfflineQueueForAgent(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) ReportingAttempt(attemptId int64, status, description string, nextOffering *int64, expireAt *int64, vars map[string]string,
	stickyDisplay bool, agentId int32) *model.AppError {

	err := app.cc.Member().AttemptResult(attemptId, status, description, nextOffering, expireAt, vars, stickyDisplay, agentId)

	if err != nil {
		return model.NewAppError("ReportingAttempt", "app.cc_member.reporting.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (app *App) RenewalAttempt(domainId, attemptId int64, renewal uint32) *model.AppError {
	err := app.cc.Member().RenewalResult(domainId, attemptId, renewal)
	if err != nil {
		return model.NewAppError("RenewalAttempt", "app.cc_member.renewal_attempt.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}
