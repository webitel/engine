package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/controller"
	"google.golang.org/grpc"
)

type API struct {
	app                   *app.App
	ctrl                  *controller.Controller
	calendar              *calendar
	skill                 *skill
	agentTeam             *agentTeam
	teamHook              *teamHook
	teamTrigger           *teamTrigger
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

	chatPlan       *chatPlanApi
	trigger        *trigger
	chatHelper     *chatHelper
	auditForm      *auditForm
	presetQuery    *presetQuery
	systemSettings *systemSettings
	webHook        *webHook
	schemaVersion  *schemaVersion
	schemaVariable *schemaVariable
	push           *push
}

func Init(a *app.App, server *grpc.Server) {
	api := &API{
		app:  a,
		ctrl: controller.NewController(a),
	}
	api.calendar = NewCalendarApi(api)
	api.skill = NewSkillApi(api)
	api.agentTeam = NewAgentTeamApi(a)
	api.teamHook = NewTeamHookApi(api)
	api.teamTrigger = NewTeamTriggerApi(api)
	api.agent = NewAgentApi(api)
	api.agentSkill = NewAgentSkillApi(api)
	api.outboundResource = NewOutboundResourceApi(a)
	api.outboundResourceGroup = NewOutboundResourceGroupApi(a)
	api.queue = NewQueueApi(a, api)
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
	api.auditForm = NewAuditFormApi(api)
	api.presetQuery = NewPresetQueryApi(api)
	api.systemSettings = NewSystemSettingsApi(api)
	api.schemaVersion = NewSchemeVersionApi(api)
	api.schemaVariable = NewSchemeVariableApi(api)
	api.webHook = NewWebHookApi(api)
	api.push = NewPushApi(api, a.Config().MinimumNumberMaskLen, a.Config().PrefixNumberMaskLen, a.Config().SuffixNumberMaskLen)

	gogrpc.RegisterCalendarServiceServer(server, api.calendar)
	gogrpc.RegisterSkillServiceServer(server, api.skill)
	gogrpc.RegisterAgentTeamServiceServer(server, api.agentTeam)
	gogrpc.RegisterTeamHookServiceServer(server, api.teamHook)
	gogrpc.RegisterTeamTriggerServiceServer(server, api.teamTrigger)
	gogrpc.RegisterAgentServiceServer(server, api.agent)
	gogrpc.RegisterAgentSkillServiceServer(server, api.agentSkill)
	gogrpc.RegisterOutboundResourceServiceServer(server, api.outboundResource)
	gogrpc.RegisterOutboundResourceGroupServiceServer(server, api.outboundResourceGroup)
	gogrpc.RegisterQueueServiceServer(server, api.queue)
	gogrpc.RegisterQueueResourcesServiceServer(server, api.queueResource)
	gogrpc.RegisterQueueSkillServiceServer(server, api.queueSkill)
	gogrpc.RegisterQueueHookServiceServer(server, api.queueHook)
	gogrpc.RegisterCommunicationTypeServiceServer(server, api.communicationType)
	gogrpc.RegisterBucketServiceServer(server, api.bucket)
	gogrpc.RegisterQueueBucketServiceServer(server, api.queueBucket)
	gogrpc.RegisterListServiceServer(server, api.list)

	gogrpc.RegisterMemberServiceServer(server, api.member)

	gogrpc.RegisterRoutingSchemaServiceServer(server, api.routingSchema)
	gogrpc.RegisterRoutingOutboundCallServiceServer(server, api.routingOutboundCall)
	gogrpc.RegisterRoutingVariableServiceServer(server, api.routingVariable)

	gogrpc.RegisterCallServiceServer(server, api.call)
	gogrpc.RegisterEmailProfileServiceServer(server, api.emailProfile)
	gogrpc.RegisterRegionServiceServer(server, api.region)
	gogrpc.RegisterAgentPauseCauseServiceServer(server, api.pauseCause)
	gogrpc.RegisterUserHelperServiceServer(server, api.userHelper)
	gogrpc.RegisterRoutingChatPlanServiceServer(server, api.chatPlan)
	gogrpc.RegisterTriggerServiceServer(server, api.trigger)

	gogrpc.RegisterChatHelperServiceServer(server, api.chatHelper)
	gogrpc.RegisterAuditFormServiceServer(server, api.auditForm)
	gogrpc.RegisterPresetQueryServiceServer(server, api.presetQuery)
	gogrpc.RegisterSystemSettingServiceServer(server, api.systemSettings)
	gogrpc.RegisterWebHookServiceServer(server, api.webHook)
	gogrpc.RegisterSchemaVersionServiceServer(server, api.schemaVersion)
	gogrpc.RegisterSchemaVariablesServiceServer(server, api.schemaVariable)
	gogrpc.RegisterPushServiceServer(server, api.push)
}
