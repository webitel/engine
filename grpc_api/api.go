package grpc_api

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/controller"
	"github.com/webitel/engine/gen/engine"
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
	quickReply   *quickReply

	chatPlan       *chatPlanApi
	trigger        *trigger
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
	api.quickReply = NewQuickReply(api)
	api.chatPlan = NewChatPlan(api)
	api.trigger = newTriggerApi(api)
	api.auditForm = NewAuditFormApi(api)
	api.presetQuery = NewPresetQueryApi(api)
	api.systemSettings = NewSystemSettingsApi(api)
	api.schemaVersion = NewSchemeVersionApi(api)
	api.schemaVariable = NewSchemeVariableApi(api)
	api.webHook = NewWebHookApi(api)
	api.push = NewPushApi(api, a.Config().MinimumNumberMaskLen, a.Config().PrefixNumberMaskLen, a.Config().SuffixNumberMaskLen)

	engine.RegisterCalendarServiceServer(server, api.calendar)
	engine.RegisterSkillServiceServer(server, api.skill)
	engine.RegisterAgentTeamServiceServer(server, api.agentTeam)
	engine.RegisterTeamHookServiceServer(server, api.teamHook)
	engine.RegisterTeamTriggerServiceServer(server, api.teamTrigger)
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
	engine.RegisterQuickRepliesServiceServer(server, api.quickReply)
	engine.RegisterRoutingChatPlanServiceServer(server, api.chatPlan)
	engine.RegisterTriggerServiceServer(server, api.trigger)

	engine.RegisterAuditFormServiceServer(server, api.auditForm)
	engine.RegisterPresetQueryServiceServer(server, api.presetQuery)
	engine.RegisterSystemSettingServiceServer(server, api.systemSettings)
	engine.RegisterWebHookServiceServer(server, api.webHook)
	engine.RegisterSchemaVersionServiceServer(server, api.schemaVersion)
	engine.RegisterSchemaVariablesServiceServer(server, api.schemaVariable)
	engine.RegisterPushServiceServer(server, api.push)
}
