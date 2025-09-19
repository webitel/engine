package grpc_api

import (
	"context"
	"strings"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type queue struct {
	*API
	app *app.App
	engine.UnsafeQueueServiceServer
}

func NewQueueApi(app *app.App, api *API) *queue {
	return &queue{app: app, API: api}
}

func (api *queue) CreateQueue(ctx context.Context, in *engine.CreateQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	queue := &model.Queue{
		Strategy:             in.Strategy,
		Enabled:              in.Enabled,
		Payload:              MarshalJsonpbToMap(in.Payload),
		Calendar:             GetLookup(in.GetCalendar()),
		Priority:             int(in.Priority),
		Name:                 in.Name,
		Variables:            in.Variables,
		Timeout:              int(in.Timeout),
		DncList:              GetLookup(in.GetDncList()),
		Ringtone:             GetLookup(in.GetRingtone()),
		SecLocateAgent:       int(in.SecLocateAgent),
		Type:                 int8(in.Type),
		Team:                 GetLookup(in.GetTeam()),
		Schema:               GetLookup(in.GetSchema()),
		DoSchema:             GetLookup(in.GetDoSchema()),
		AfterSchema:          GetLookup(in.GetAfterSchema()),
		Description:          in.Description,
		StickyAgent:          in.StickyAgent,
		Processing:           in.Processing,
		ProcessingSec:        in.ProcessingSec,
		ProcessingRenewalSec: in.ProcessingRenewalSec,
		FormSchema:           GetLookup(in.GetFormSchema()),
		Grantee:              GetLookup(in.GetGrantee()),
		Tags:                 tagsToStrings(in.GetTags()),
		TaskProcessing: &model.QueueTaskProcessing{ //?rewrite better
			ProlongationOptions: &model.QueueTaskProcessingProlongationOptions{},
		},
	}

	if in.TaskProcessing != nil {
		queue.Processing = in.TaskProcessing.Enabled
		queue.ProcessingSec = in.TaskProcessing.Sec
		queue.ProcessingRenewalSec = in.TaskProcessing.RenewalSec
		queue.FormSchema = GetLookup(in.TaskProcessing.GetFormSchema())

		if in.TaskProcessing.ProlongationOptions != nil {
			queue.TaskProcessing.ProlongationOptions.ProlongationEnabled = in.TaskProcessing.ProlongationOptions.Enabled
			queue.TaskProcessing.ProlongationOptions.IsTimeoutRetry = in.TaskProcessing.ProlongationOptions.IsTimeoutRetry
			queue.TaskProcessing.ProlongationOptions.ProlongationTimeSec = in.TaskProcessing.ProlongationOptions.ProlongationTimeSec
			queue.TaskProcessing.ProlongationOptions.RepeatsNumber = in.TaskProcessing.ProlongationOptions.RepeatsNumber
		}
	}

	queue, err = api.ctrl.CreateQueue(ctx, session, queue)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
}

func (api *queue) SearchQueue(ctx context.Context, in *engine.SearchQueueRequest) (*engine.ListQueue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Queue
	var endList bool
	req := &model.SearchQueue{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:     in.Id,
		Types:   in.Type,
		TeamIds: in.TeamId,
		Tags:    in.GetTags(),
	}

	if in.Enabled {
		req.Enabled = &in.Enabled
	}

	list, endList, err = api.ctrl.SearchQueue(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Queue, 0, len(list))
	for _, v := range list {
		items = append(items, transformQueue(v))
	}

	return &engine.ListQueue{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *queue) ReadQueue(ctx context.Context, in *engine.ReadQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	queue, err := api.ctrl.GetQueue(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
}

func (api *queue) PatchQueue(ctx context.Context, in *engine.PatchQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	patch := &model.QueuePatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "strategy":
			patch.Strategy = model.NewString(in.Strategy)
		case "enabled":
			patch.Enabled = model.NewBool(in.Enabled)
		case "calendar.id":
			patch.Calendar = GetLookup(in.Calendar)
		case "priority":
			patch.Priority = model.NewInt(int(in.Priority))
		case "name":
			patch.Name = model.NewString(in.Name)
		case "dnc_list.id":
			patch.DncList = GetLookup(in.DncList)
		case "team.id":
			patch.Team = GetLookup(in.Team)
		case "schema.id":
			patch.Schema = GetLookup(in.Schema)
		case "ringtone.id":
			patch.Ringtone = GetLookup(in.Ringtone)
		case "do_schema.id":
			patch.DoSchema = GetLookup(in.DoSchema)
		case "after_schema.id":
			patch.AfterSchema = GetLookup(in.AfterSchema)
		case "form_schema.id":
			patch.FormSchema = GetLookup(in.FormSchema)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "sticky_agent":
			patch.StickyAgent = model.NewBool(in.StickyAgent)
		case "processing":
			patch.Processing = &in.Processing
		case "task_processing.enabled":
			patch.Processing = &in.GetTaskProcessing().Enabled
		case "processing_sec":
			patch.ProcessingSec = &in.ProcessingSec
		case "task_processing.sec":
			patch.ProcessingSec = &in.GetTaskProcessing().Sec
		case "task_processing.renewal_sec":
			patch.ProcessingRenewalSec = &in.GetTaskProcessing().RenewalSec
		case "processing_renewal_sec":
			patch.ProcessingRenewalSec = &in.ProcessingRenewalSec
		case "grantee.id":
			patch.Grantee = GetLookup(in.Grantee)
		case "tags":
			patch.Tags = tagsToStrings(in.Tags)
		case "task_processing.prolongation_options.enabled":
			patch.ProlongationEnabled = &in.GetTaskProcessing().GetProlongationOptions().Enabled
		case "task_processing.prolongation_options.repeats_number":
			patch.RepeatsNumber = &in.GetTaskProcessing().GetProlongationOptions().RepeatsNumber
		case "task_processing.prolongation_options.prolongation_time_sec":
			patch.ProlongationTimeSec = &in.GetTaskProcessing().GetProlongationOptions().ProlongationTimeSec
		case "task_processing.prolongation_options.prolongation_is_timeout_retry":
			patch.IsTimeoutRetry = &in.GetTaskProcessing().GetProlongationOptions().IsTimeoutRetry
		default:
			if patch.Variables == nil && strings.HasPrefix(v, "variables.") {
				patch.Variables = in.Variables
			} else if patch.Payload == nil && strings.HasPrefix(v, "payload.") {
				patch.Payload = MarshalJsonpbToMap(in.Payload)
			}
		}
	}

	queue, err := api.ctrl.PatchQueue(ctx, session, in.GetId(), patch)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
}

func (api *queue) UpdateQueue(ctx context.Context, in *engine.UpdateQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	queue := &model.Queue{
		DomainRecord: model.DomainRecord{
			Id: in.Id,
		},
		Strategy:             in.Strategy,
		Enabled:              in.Enabled,
		Payload:              MarshalJsonpbToMap(in.Payload),
		Calendar:             GetLookup(in.Calendar),
		Priority:             int(in.Priority),
		Name:                 in.Name,
		Variables:            in.Variables,
		Timeout:              int(in.Timeout),
		DncList:              GetLookup(in.DncList),
		SecLocateAgent:       int(in.SecLocateAgent),
		Type:                 int8(in.Type),
		Team:                 GetLookup(in.Team),
		Schema:               GetLookup(in.Schema),
		Ringtone:             GetLookup(in.Ringtone),
		DoSchema:             GetLookup(in.DoSchema),
		AfterSchema:          GetLookup(in.AfterSchema),
		FormSchema:           GetLookup(in.GetFormSchema()),
		Description:          in.Description,
		StickyAgent:          in.StickyAgent,
		Processing:           in.Processing,
		ProcessingSec:        in.ProcessingSec,
		ProcessingRenewalSec: in.ProcessingRenewalSec,
		Grantee:              GetLookup(in.GetGrantee()),
		Tags:                 tagsToStrings(in.GetTags()),
		TaskProcessing: &model.QueueTaskProcessing{
			ProlongationOptions: &model.QueueTaskProcessingProlongationOptions{},
		},
	}

	if in.TaskProcessing != nil {
		queue.Processing = in.TaskProcessing.Enabled
		queue.ProcessingSec = in.TaskProcessing.Sec
		queue.ProcessingRenewalSec = in.TaskProcessing.RenewalSec
		queue.FormSchema = GetLookup(in.TaskProcessing.GetFormSchema())

		if in.TaskProcessing.ProlongationOptions != nil {
			queue.TaskProcessing.ProlongationOptions.ProlongationEnabled = in.TaskProcessing.ProlongationOptions.Enabled
			queue.TaskProcessing.ProlongationOptions.IsTimeoutRetry = in.TaskProcessing.ProlongationOptions.IsTimeoutRetry
			queue.TaskProcessing.ProlongationOptions.ProlongationTimeSec = in.TaskProcessing.ProlongationOptions.ProlongationTimeSec
			queue.TaskProcessing.ProlongationOptions.RepeatsNumber = in.TaskProcessing.ProlongationOptions.RepeatsNumber
		}
	}

	queue, err = api.ctrl.UpdateQueue(ctx, session, queue)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
}

func (api *queue) DeleteQueue(ctx context.Context, in *engine.DeleteQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var queue *model.Queue
	queue, err = api.ctrl.DeleteQueue(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
}

func (api *queue) SearchQueueReportGeneral(ctx context.Context, in *engine.SearchQueueReportGeneralRequest) (*engine.ListReportGeneral, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetJoinedAt() == nil {
		return nil, model.NewBadRequestError("grpc.queue.report.general", "filter joined_at is required")
	}

	var report *model.QueueReportGeneralAgg
	var list []*model.QueueReportGeneral
	var endList bool
	req := &model.SearchQueueReportGeneral{
		ListRequest: model.ListRequest{
			DomainId: session.Domain(in.DomainId),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Q:        in.GetQ(),
			Fields:   in.Fields,
			Sort:     in.Sort,
		},
		JoinedAt: model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		},
		QueueIds: in.QueueId,
		TeamIds:  in.TeamId,
		Types:    in.GetType(),
	}

	report, endList, err = api.ctrl.QueueReportGeneral(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.QueueReportGeneral, 0, len(list))
	for _, v := range report.Items {
		items = append(items, toEngineQueueReportGeneral(v))
	}
	return &engine.ListReportGeneral{
		Next:  !endList,
		Items: items,
		Aggs: &engine.QueueReportGeneralAgentStatus{
			Online:  report.Aggs.Online,
			Pause:   report.Aggs.Pause,
			Offline: report.Aggs.Offline,
			Free:    report.Aggs.Free,
			Total:   report.Aggs.Total,
		},
	}, nil
}

func (api *queue) SearchQueueTags(ctx context.Context, in *engine.SearchQueueTagsRequest) (*engine.ListTags, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Tag
	var endList bool

	req := &model.ListRequest{
		Q:       in.GetQ(),
		Page:    int(in.GetPage()),
		PerPage: int(in.GetSize()),
		Fields:  in.Fields,
		Sort:    in.Sort,
	}

	list, endList, err = api.ctrl.SearchQueueTags(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Tag, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.Tag{
			Name: v.Name,
		})
	}
	return &engine.ListTags{
		Next:  !endList,
		Items: items,
	}, nil
}

func toEngineQueueReportGeneral(src *model.QueueReportGeneral) *engine.QueueReportGeneral {
	return &engine.QueueReportGeneral{
		Queue: GetProtoLookup(&src.Queue),
		AgentStatus: &engine.QueueReportGeneralAgentStatus{
			Online:  src.AgentStatus.Online,
			Pause:   src.AgentStatus.Pause,
			Offline: src.AgentStatus.Offline,
			Free:    src.AgentStatus.Free,
			Total:   src.AgentStatus.Total,
		},
		Team:        GetProtoLookup(src.Team),
		Missed:      src.Missed,
		Processed:   src.Processed,
		Waiting:     src.Waiting,
		Count:       src.Count,
		Transferred: src.Transferred,
		Abandoned:   src.Abandoned,
		Attempts:    src.Attempts,
		SumBillSec:  src.SumBillSec,
		AvgWrapSec:  src.AvgWrapSec,
		AvgAwtSec:   src.AvgAwtSec,
		AvgAsaSec:   src.AvgAsaSec,
		AvgAhtSec:   src.AvgAhtSec,
		Bridged:     src.Bridged,
		Sl20:        src.Sl20,
		Sl30:        src.Sl30,
	}
}

func transformQueue(src *model.Queue) *engine.Queue {
	q := &engine.Queue{
		Id:                   src.Id,
		DomainId:             src.DomainId,
		CreatedAt:            src.CreatedAt,
		CreatedBy:            GetProtoLookup(src.CreatedBy),
		UpdatedAt:            src.UpdatedAt,
		UpdatedBy:            GetProtoLookup(src.UpdatedBy),
		Strategy:             src.Strategy,
		Enabled:              src.Enabled,
		Payload:              UnmarshalJsonpb(src.Payload.ToSafeBytes()),
		Calendar:             GetProtoLookup(src.Calendar),
		Priority:             int32(src.Priority),
		Name:                 src.Name,
		Variables:            src.Variables,
		Timeout:              int32(src.Timeout),
		DncList:              GetProtoLookup(src.DncList),
		SecLocateAgent:       int32(src.SecLocateAgent),
		Type:                 int32(src.Type),
		Team:                 GetProtoLookup(src.Team),
		Schema:               GetProtoLookup(src.Schema),
		Ringtone:             GetProtoLookup(src.Ringtone),
		DoSchema:             GetProtoLookup(src.DoSchema),
		AfterSchema:          GetProtoLookup(src.AfterSchema),
		FormSchema:           GetProtoLookup(src.FormSchema),
		Description:          src.Description,
		Count:                int32(src.Count),
		Waiting:              int32(src.Waiting),
		Active:               int32(src.Active),
		StickyAgent:          src.StickyAgent,
		Processing:           src.Processing,
		ProcessingSec:        src.ProcessingSec,
		ProcessingRenewalSec: src.ProcessingRenewalSec,
		Grantee:              GetProtoLookup(src.Grantee),
		Tags:                 stringsToTags(src.Tags),
	}

	if src.TaskProcessing != nil {
		q.TaskProcessing = &engine.TaskProcessing{
			Enabled:    src.TaskProcessing.Enabled,
			FormSchema: GetProtoLookup(src.TaskProcessing.FormSchema),
			Sec:        src.TaskProcessing.Sec,
			RenewalSec: src.TaskProcessing.RenewalSec,
		}

		if src.TaskProcessing.ProlongationOptions != nil {
			q.TaskProcessing.ProlongationOptions = &engine.TaskProcessingProlongationOptions{
				Enabled:             src.TaskProcessing.ProlongationOptions.ProlongationEnabled,
				RepeatsNumber:       src.TaskProcessing.ProlongationOptions.RepeatsNumber,
				ProlongationTimeSec: src.TaskProcessing.ProlongationOptions.ProlongationTimeSec,
				IsTimeoutRetry:      src.TaskProcessing.ProlongationOptions.IsTimeoutRetry,
			}
		}
	}

	return q
}

func (api *queue) SetQueuesGlobalState(ctx context.Context, in *engine.SetQueuesGlobalStateRequest) (*engine.SetQueuesGlobalStateResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := api.ctrl.SetQueuesGlobalState(ctx, session, in.Enabled)
	if err != nil {
		return nil, err
	}

	return &engine.SetQueuesGlobalStateResponse{
		Count: rowsAffected,
	}, nil
}

func (api *queue) GetQueuesGlobalState(ctx context.Context, in *engine.GetQueuesGlobalStateRequest) (*engine.GetQueuesGlobalStateResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	res, err := api.ctrl.GetQueuesGlobalState(ctx, session)
	if err != nil {
		return nil, err
	}

	return &engine.GetQueuesGlobalStateResponse{
		IsAllEnabled: res,
	}, nil
}
