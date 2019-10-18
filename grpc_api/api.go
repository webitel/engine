package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"google.golang.org/grpc"
)

type API struct {
	app                   *app.App
	calendar              *calendar
	skill                 *skill
	agentTeam             *agentTeam
	agent                 *agent
	agentSkill            *agentSkill
	outboundResource      *outboundResource
	outboundResourceGroup *outboundResourceGroup
	queue                 *queue
	queueRouting          *queueRouting
	supervisorInTeam      *supervisorInTeam
	communicationType     *communicationType

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
	api.outboundResourceGroup = NewOutboundResourceGroupApi(a)
	api.queue = NewQueueApi(a)
	api.queueRouting = NewQueueRoutingApi(a)
	api.supervisorInTeam = NewSupervisorInTeamApi(a)

	api.routingScheme = NewRoutingSchemeApi(a)
	api.routingOutboundCall = NewRoutingOutboundCallApi(a)
	api.routingVariable = NewRoutingVariableApi(a)
	api.communicationType = NewCommunicationTypeApi(a)

	engine.RegisterCalendarServiceServer(server, api.calendar)
	engine.RegisterSkillServiceServer(server, api.skill)
	engine.RegisterAgentTeamServiceServer(server, api.agentTeam)
	engine.RegisterAgentServiceServer(server, api.agent)
	engine.RegisterAgentSkillServiceServer(server, api.agentSkill)
	engine.RegisterResourceTeamServiceServer(server, api.resourceTeam)
	engine.RegisterOutboundResourceServiceServer(server, api.outboundResource)
	engine.RegisterOutboundResourceGroupServiceServer(server, api.outboundResourceGroup)
	engine.RegisterQueueServiceServer(server, api.queue)
	engine.RegisterQueueRoutingServiceServer(server, api.queueRouting)
	engine.RegisterSupervisorInTeamServiceServer(server, api.supervisorInTeam)
	engine.RegisterCommunicationTypeServiceServer(server, api.communicationType)

	engine.RegisterRoutingSchemeServiceServer(server, api.routingScheme)
	engine.RegisterRoutingOutboundCallServiceServer(server, api.routingOutboundCall)
	engine.RegisterRoutingVariableServiceServer(server, api.routingVariable)
}
