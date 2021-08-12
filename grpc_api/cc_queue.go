package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"net/http"
	"strings"
)

type queue struct {
	app *app.App
}

func NewQueueApi(app *app.App) *queue {
	return &queue{app: app}
}

func (api *queue) CreateQueue(ctx context.Context, in *engine.CreateQueueRequest) (*engine.Queue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	queue := &model.Queue{
		DomainRecord: model.DomainRecord{
			Id:        0,
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Strategy: in.Strategy,
		Enabled:  in.Enabled,
		Payload:  MarshalJsonpb(in.Payload),
		Calendar: model.Lookup{
			Id:   int(in.GetCalendar().GetId()),
			Name: in.GetCalendar().GetName(),
		},
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
	}

	if err = queue.IsValid(); err != nil {
		return nil, err
	}

	queue, err = api.app.CreateQueue(queue)
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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
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
		Ids:   in.Id,
		Types: in.Type,
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.GetQueuePageByGroups(session.Domain(0), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetQueuePage(session.Domain(0), req)
	}

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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var queue *model.Queue

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	queue, err = api.app.GetQueueById(session.Domain(in.DomainId), in.Id)

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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var queue *model.Queue
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
		case "description":
			patch.Description = model.NewString(in.Description)
		case "sticky_agent":
			patch.StickyAgent = model.NewBool(in.StickyAgent)
		case "processing":
			patch.Processing = &in.Processing
		case "processing_sec":
			patch.ProcessingSec = &in.ProcessingSec
		case "processing_renewal_sec":
			patch.ProcessingRenewalSec = &in.ProcessingRenewalSec
		default:
			if patch.Variables == nil && strings.HasPrefix(v, "variables.") {
				patch.Variables = in.Variables
			} else if patch.Payload == nil && strings.HasPrefix(v, "payload.") {
				patch.Payload = MarshalJsonpb(in.Payload)
			}
		}
	}

	queue, err = api.app.PatchQueue(session.Domain(in.GetDomainId()), in.GetId(), patch)

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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var queue *model.Queue

	queue, err = api.app.UpdateQueue(&model.Queue{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Strategy: in.Strategy,
		Enabled:  in.Enabled,
		Payload:  MarshalJsonpb(in.Payload),
		Calendar: model.Lookup{
			Id: int(in.GetCalendar().GetId()),
		},
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
		Description:          in.Description,
		StickyAgent:          in.StickyAgent,
		Processing:           in.Processing,
		ProcessingSec:        in.ProcessingSec,
		ProcessingRenewalSec: in.ProcessingRenewalSec,
	})

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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var queue *model.Queue
	queue, err = api.app.RemoveQueue(session.Domain(in.DomainId), in.Id)
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

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if in.GetJoinedAt() == nil {
		return nil, model.NewAppError("GRPC.SearchQueueReportGeneral", "grpc.queue.report.general", nil, "filter joined_at is required", http.StatusBadRequest)
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

	report, endList, err = api.app.GetQueueReportGeneral(session.Domain(in.DomainId), session.UserId, session.RoleIds, auth_manager.PERMISSION_ACCESS_READ, req)
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
	return &engine.Queue{
		Id:        src.Id,
		DomainId:  src.DomainId,
		CreatedAt: src.CreatedAt,
		CreatedBy: &engine.Lookup{
			Id:   int64(src.CreatedBy.Id),
			Name: src.CreatedBy.Name,
		},
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: &engine.Lookup{
			Id:   int64(src.UpdatedBy.Id),
			Name: src.UpdatedBy.Name,
		},
		Strategy: src.Strategy,
		Enabled:  src.Enabled,
		Payload:  UnmarshalJsonpb(src.Payload),
		Calendar: &engine.Lookup{
			Id:   int64(src.Calendar.Id),
			Name: src.Calendar.Name,
		},
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
		Description:          src.Description,
		Count:                int32(src.Count),
		Waiting:              int32(src.Waiting),
		Active:               int32(src.Active),
		StickyAgent:          src.StickyAgent,
		Processing:           src.Processing,
		ProcessingSec:        src.ProcessingSec,
		ProcessingRenewalSec: src.ProcessingRenewalSec,
	}
}
