package app

import "github.com/webitel/engine/model"

func (app *App) CreateSupervisorInTeam(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError) {
	return app.Store.SupervisorTeam().Create(supervisor)
}

func (app *App) GetSupervisorsPage(domainId, teamId int64, page, perPage int) ([]*model.SupervisorInTeam, *model.AppError) {
	return app.Store.SupervisorTeam().GetAllPage(domainId, teamId, page*perPage, perPage)
}

func (app *App) GetSupervisorsInTeam(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError) {
	return app.Store.SupervisorTeam().Get(domainId, teamId, id)
}

func (app *App) UpdateSupervisorsInTeam(domainId int64, supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError) {
	oldSupervisor, err := app.Store.SupervisorTeam().Get(domainId, supervisor.TeamId, supervisor.Id)

	if err != nil {
		return nil, err
	}

	oldSupervisor.Agent = supervisor.Agent

	_, err = app.Store.SupervisorTeam().Update(oldSupervisor)
	if err != nil {
		return nil, err
	}

	return oldSupervisor, nil
}

func (app *App) RemoveSupervisorsInTeam(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError) {
	supervisor, err := app.Store.SupervisorTeam().Get(domainId, teamId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.SupervisorTeam().Delete(teamId, id)
	if err != nil {
		return nil, err
	}
	return supervisor, nil
}
