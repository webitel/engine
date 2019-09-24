package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"google.golang.org/grpc"
)

type API struct {
	app                 *app.App
	calendar            *calendar
	skill               *skill
	agentTeam           *agentTeam
	agent               *agent
	agentSkill          *agentSkill
	routingScheme       *routingScheme
	routingOutboundCall *routingOutboundCall
	routingVariable     *routingVariable
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{app: a}
	api.calendar = NewCalendarApi(a)
	api.skill = NewSkillApi(a)
	api.agentTeam = NewAgentTeamApi(a)
	api.agent = NewAgentApi(a)
	api.agentSkill = NewAgentSkillApi(a)
	api.routingScheme = NewRoutingSchemeApi(a)
	api.routingOutboundCall = NewRoutingOutboundCallApi(a)
	api.routingVariable = NewRoutingVariableApi(a)

	engine.RegisterCalendarApiServer(server, api.calendar)
	engine.RegisterSkillApiServer(server, api.skill)
	engine.RegisterAgentTeamApiServer(server, api.agentTeam)
	engine.RegisterAgentApiServer(server, api.agent)
	engine.RegisterAgentSkillApiServer(server, api.agentSkill)
	engine.RegisterRoutingSchemeApiServer(server, api.routingScheme)
	engine.RegisterRoutingOutboundCallApiServer(server, api.routingOutboundCall)
	engine.RegisterRoutingVariableApiServer(server, api.routingVariable)
}
