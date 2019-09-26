package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"google.golang.org/grpc"
)

type API struct {
	app              *app.App
	calendar         *calendar
	skill            *skill
	agentTeam        *agentTeam
	agent            *agent
	agentSkill       *agentSkill
	outboundResource *outboundResource
	queue            *queue
	supervisorInTeam *supervisorInTeam

	routingScheme       *routingScheme
	routingOutboundCall *routingOutboundCall
	routingVariable     *routingVariable
	resourceTeam        *resourceTeam
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{app: a}
	api.calendar = NewCalendarApi(a)
	api.skill = NewSkillApi(a)
	api.agentTeam = NewAgentTeamApi(a)
	api.agent = NewAgentApi(a)
	api.agentSkill = NewAgentSkillApi(a)
	api.resourceTeam = NewResourceTeamApi(a)
	api.outboundResource = NewOutboundResourceApi(a)
	api.queue = NewQueueApi(a)
	api.supervisorInTeam = NewSupervisorInTeamApi(a)

	api.routingScheme = NewRoutingSchemeApi(a)
	api.routingOutboundCall = NewRoutingOutboundCallApi(a)
	api.routingVariable = NewRoutingVariableApi(a)

	engine.RegisterCalendarApiServer(server, api.calendar)
	engine.RegisterSkillApiServer(server, api.skill)
	engine.RegisterAgentTeamApiServer(server, api.agentTeam)
	engine.RegisterAgentApiServer(server, api.agent)
	engine.RegisterAgentSkillApiServer(server, api.agentSkill)
	engine.RegisterResourceTeamApiServer(server, api.resourceTeam)
	engine.RegisterOutboundResourceApiServer(server, api.outboundResource)
	engine.RegisterQueueApiServer(server, api.queue)
	engine.RegisterSupervisorInTeamApiServer(server, api.supervisorInTeam)

	engine.RegisterRoutingSchemeApiServer(server, api.routingScheme)
	engine.RegisterRoutingOutboundCallApiServer(server, api.routingOutboundCall)
	engine.RegisterRoutingVariableApiServer(server, api.routingVariable)
}
