package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/controller"
	"github.com/webitel/engine/grpc_api/engine"
	"google.golang.org/grpc"
)

type API struct {
	app                   *app.App
	ctrl                  *controller.Controller
	calendar              *calendar
	skill                 *skill
	agentTeam             *agentTeam
	agent                 *agent
	agentSkill            *agentSkill
	outboundResource      *outboundResource
	outboundResourceGroup *outboundResourceGroup
	queue                 *queue
	queueResource         *queueResource
	queueRouting          *queueRouting
	supervisorInTeam      *supervisorInTeam
	communicationType     *communicationType
	member                *member
	bucket                *bucket
	queueBucket           *queueBucket
	list                  *list

	routingSchema       *routingSchema
	routingOutboundCall *routingOutboundCall
	routingVariable     *routingVariable
	resourceTeam        *resourceTeam

	call *call

	emailProfile *emailProfile
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{
		app:  a,
		ctrl: controller.NewController(a),
	}
	api.calendar = NewCalendarApi(api)
	api.skill = NewSkillApi(a)
	api.agentTeam = NewAgentTeamApi(a)
	api.agent = NewAgentApi(api)
	api.agentSkill = NewAgentSkillApi(a)
	api.resourceTeam = NewResourceTeamApi(a)
	api.outboundResource = NewOutboundResourceApi(a)
	api.outboundResourceGroup = NewOutboundResourceGroupApi(a)
	api.queue = NewQueueApi(a)
	api.queueResource = NewQueueResourceApi(a)
	api.queueRouting = NewQueueRoutingApi(a)
	api.supervisorInTeam = NewSupervisorInTeamApi(a)

	api.routingSchema = NewRoutingSchemaApi(a)
	api.routingOutboundCall = NewRoutingOutboundCallApi(api)
	api.routingVariable = NewRoutingVariableApi(a)
	api.communicationType = NewCommunicationTypeApi(a)
	api.bucket = NewBucketApi(a)
	api.queueBucket = NewQueueBucketApi(a)
	api.list = NewListApi(a)

	api.member = NewMemberApi(api)

	api.call = NewCallApi(api)
	api.emailProfile = NewEmailProfileApi(api)

	engine.RegisterCalendarServiceServer(server, api.calendar)
	engine.RegisterSkillServiceServer(server, api.skill)
	engine.RegisterAgentTeamServiceServer(server, api.agentTeam)
	engine.RegisterAgentServiceServer(server, api.agent)
	engine.RegisterAgentSkillServiceServer(server, api.agentSkill)
	engine.RegisterResourceTeamServiceServer(server, api.resourceTeam)
	engine.RegisterOutboundResourceServiceServer(server, api.outboundResource)
	engine.RegisterOutboundResourceGroupServiceServer(server, api.outboundResourceGroup)
	engine.RegisterQueueServiceServer(server, api.queue)
	engine.RegisterQueueResourcesServiceServer(server, api.queueResource)
	engine.RegisterQueueRoutingServiceServer(server, api.queueRouting)
	engine.RegisterSupervisorInTeamServiceServer(server, api.supervisorInTeam)
	engine.RegisterCommunicationTypeServiceServer(server, api.communicationType)
	engine.RegisterBucketServiceServer(server, api.bucket)
	engine.RegisterQueueBucketServiceServer(server, api.queueBucket)
	engine.RegisterListServiceServer(server, api.list)

	engine.RegisterMemberServiceServer(server, api.member)

	engine.RegisterRoutingSchemaServiceServer(server, api.routingSchema)
	engine.RegisterRoutingOutboundCallServiceServer(server, api.routingOutboundCall)
	engine.RegisterRoutingVariableServiceServer(server, api.routingVariable)

	engine.RegisterCallServiceServer(server, api.call)
	engine.RegisterEmailProfileServiceServer(server, api.emailProfile)
}
