package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/controller"
	"github.com/webitel/protos/engine"
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
	queueSkill            *queueSkill
	queueHook             *queueHook
	communicationType     *communicationType
	member                *member
	bucket                *bucket
	queueBucket           *queueBucket
	list                  *list

	routingSchema       *routingSchema
	routingOutboundCall *routingOutboundCall
	routingVariable     *routingVariable

	call *call

	emailProfile *emailProfile
	region       *region
	pauseCause   *pauseCause
	userHelper   *userHelper

	chatPlan   *chatPlanApi
	trigger    *trigger
	chatHelper *chatHelper
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{
		app:  a,
		ctrl: controller.NewController(a),
	}
	api.calendar = NewCalendarApi(api)
	api.skill = NewSkillApi(api)
	api.agentTeam = NewAgentTeamApi(a)
	api.agent = NewAgentApi(api)
	api.agentSkill = NewAgentSkillApi(a)
	api.outboundResource = NewOutboundResourceApi(a)
	api.outboundResourceGroup = NewOutboundResourceGroupApi(a)
	api.queue = NewQueueApi(a)
	api.queueResource = NewQueueResourceApi(a)
	api.queueSkill = NewQueueSkill(api)
	api.queueHook = NewQueueHookApi(api)

	api.routingSchema = NewRoutingSchemaApi(api)
	api.routingOutboundCall = NewRoutingOutboundCallApi(api)
	api.routingVariable = NewRoutingVariableApi(a)
	api.communicationType = NewCommunicationTypeApi(api)
	api.bucket = NewBucketApi(a)
	api.queueBucket = NewQueueBucketApi(a)
	api.list = NewListApi(a)

	api.member = NewMemberApi(api)

	api.call = NewCallApi(api, a.Config().MinimumNumberMaskLen, a.Config().PrefixNumberMaskLen, a.Config().SuffixNumberMaskLen)
	api.emailProfile = NewEmailProfileApi(api)
	api.region = NewRegionApi(api)
	api.pauseCause = NewPauseCause(api)
	api.userHelper = NewUserHelperApi(api)
	api.chatPlan = NewChatPlan(api)
	api.trigger = NewTriggerApi(api)
	api.chatHelper = NewChatHelperApi(api)

	engine.RegisterCalendarServiceServer(server, api.calendar)
	engine.RegisterSkillServiceServer(server, api.skill)
	engine.RegisterAgentTeamServiceServer(server, api.agentTeam)
	engine.RegisterAgentServiceServer(server, api.agent)
	engine.RegisterAgentSkillServiceServer(server, api.agentSkill)
	engine.RegisterOutboundResourceServiceServer(server, api.outboundResource)
	engine.RegisterOutboundResourceGroupServiceServer(server, api.outboundResourceGroup)
	engine.RegisterQueueServiceServer(server, api.queue)
	engine.RegisterQueueResourcesServiceServer(server, api.queueResource)
	engine.RegisterQueueSkillServiceServer(server, api.queueSkill)
	engine.RegisterQueueHookServiceServer(server, api.queueHook)
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
	engine.RegisterRegionServiceServer(server, api.region)
	engine.RegisterAgentPauseCauseServiceServer(server, api.pauseCause)
	engine.RegisterUserHelperServiceServer(server, api.userHelper)
	engine.RegisterRoutingChatPlanServiceServer(server, api.chatPlan)
	engine.RegisterTriggerServiceServer(server, api.trigger)

	engine.RegisterChatHelperServiceServer(server, api.chatHelper)
}
