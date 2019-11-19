package app

import "github.com/webitel/engine/model"

func (app *App) CreateMember(member *model.Member) (*model.Member, *model.AppError) {
	return app.Store.Member().Create(member)
}

func (app *App) BulkCreateMember(domainId, queueId int64, members []*model.Member) ([]int64, *model.AppError) {
	_, err := app.GetQueueById(domainId, queueId)
	if err != nil {
		return nil, err
	}
	return app.Store.Member().BulkCreate(queueId, members)
}

func (app *App) GetMemberPage(domainId, queueId int64, page, perPage int) ([]*model.Member, *model.AppError) {
	return app.Store.Member().GetAllPage(domainId, queueId, page*perPage, perPage)
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
	oldMember.Skills = member.Skills

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
