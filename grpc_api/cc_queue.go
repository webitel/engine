package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
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
		Priority:       int(in.Priority),
		Name:           in.Name,
		Variables:      in.Variables,
		Timeout:        int(in.Timeout),
		DncList:        GetLookup(in.GetDncList()),
		SecLocateAgent: int(in.SecLocateAgent),
		Type:           int8(in.Type),
		Team:           GetLookup(in.GetTeam()),
		Description:    in.Description,
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var list []*model.Queue

	if permission.Rbac {
		list, err = api.app.GetQueuePageByGroups(session.Domain(in.DomainId), session.RoleIds, int(in.Page), int(in.Size))
	} else {
		list, err = api.app.GetQueuePage(session.Domain(in.DomainId), int(in.Page), int(in.Size))
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Queue, 0, len(list))
	for _, v := range list {
		items = append(items, transformQueue(v))
	}
	return &engine.ListQueue{
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var queue *model.Queue

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
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
		case "payload":
			patch.Payload = MarshalJsonpb(in.Payload)
		case "calendar":
			patch.Calendar = GetLookup(in.Calendar)
		case "priority":
			patch.Priority = model.NewInt(int(in.Priority))
		case "name":
			patch.Name = model.NewString(in.Name)
		case "variables":
			patch.Variables = in.Variables
		case "timeout":
			patch.Timeout = model.NewInt(int(in.Timeout))
		case "dnc_list":
			patch.DncList = GetLookup(in.DncList)
		case "sec_locate_agent":
			patch.SecLocateAgent = model.NewInt(int(in.SecLocateAgent))
		case "team":
			patch.Team = GetLookup(in.Team)
		case "description":
			patch.Description = model.NewString(in.Description)
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
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
		Priority:       int(in.Priority),
		Name:           in.Name,
		Variables:      in.Variables,
		Timeout:        int(in.Timeout),
		DncList:        GetLookup(in.DncList),
		SecLocateAgent: int(in.SecLocateAgent),
		Type:           int8(in.Type),
		Team:           GetLookup(in.Team),
		Description:    in.Description,
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_DELETE)
		}
	}

	var queue *model.Queue
	queue, err = api.app.RemoveQueue(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformQueue(queue), nil
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
		Priority:       int32(src.Priority),
		Name:           src.Name,
		Variables:      src.Variables,
		Timeout:        int32(src.Timeout),
		DncList:        GetProtoLookup(src.DncList),
		SecLocateAgent: int32(src.SecLocateAgent),
		Type:           int32(src.Type),
		Team:           GetProtoLookup(src.Team),
		Description:    src.Description,
	}
}
