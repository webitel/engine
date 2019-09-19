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
	routingScheme       *routingScheme
	routingInboundCall  *routingInboundCall
	routingOutboundCall *routingOutboundCall
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{app: a}
	api.calendar = NewCalendarApi(a)
	api.skill = NewSkillApi(a)
	api.agentTeam = NewAgentTeamApi(a)
	api.agent = NewAgentApi(a)
	api.routingScheme = NewRoutingSchemeApi(a)
	api.routingInboundCall = NewRoutingInboundCallApi(a)
	api.routingOutboundCall = NewRoutingOutboundCallApi(a)

	engine.RegisterCalendarApiServer(server, api.calendar)
	engine.RegisterSkillApiServer(server, api.skill)
	engine.RegisterAgentTeamApiServer(server, api.agentTeam)
	engine.RegisterAgentApiServer(server, api.agent)
	engine.RegisterRoutingSchemeApiServer(server, api.routingScheme)
	engine.RegisterRoutingInboundCallApiServer(server, api.routingInboundCall)
	engine.RegisterRoutingOutboundCallApiServer(server, api.routingOutboundCall)
}
