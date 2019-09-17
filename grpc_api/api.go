package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"google.golang.org/grpc"
)

type API struct {
	app       *app.App
	calendar  *calendar
	skill     *skill
	agentTeam *agentTeam
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{app: a}
	api.calendar = NewCalendarApi(a)
	api.skill = NewSkillApi(a)
	api.agentTeam = NewAgentTeamApi(a)

	engine.RegisterCalendarApiServer(server, api.calendar)
	engine.RegisterSkillApiServer(server, api.skill)
	engine.RegisterAgentTeamApiServer(server, api.agentTeam)
}
