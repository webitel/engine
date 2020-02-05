package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type outboundResourceGroup struct {
	app *app.App
}

func NewOutboundResourceGroupApi(app *app.App) *outboundResourceGroup {
	return &outboundResourceGroup{app: app}
}

func (api *outboundResourceGroup) CreateOutboundResourceGroup(ctx context.Context, in *engine.CreateOutboundResourceGroupRequest) (*engine.OutboundResourceGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	group := &model.OutboundResourceGroup{
		DomainRecord: model.DomainRecord{
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
		Name:        in.Name,
		Strategy:    in.GetStrategy(),
		Description: in.GetDescription(),
		Communication: model.Lookup{
			Id: int(in.GetCommunication().GetId()),
		},
		Time: toModelOutboundResourceGroupTime(in.GetTime()),
	}

	if err = group.IsValid(); err != nil {
		return nil, err
	}

	if group, err = api.app.CreateOutboundResourceGroup(group); err != nil {
		return nil, err
	} else {
		return toEngineOutboundResourceGroup(group), nil
	}
}

func (api *outboundResourceGroup) SearchOutboundResourceGroup(ctx context.Context, in *engine.SearchOutboundResourceGroupRequest) (*engine.ListOutboundResourceGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.OutboundResourceGroup
	var endList bool
	req := &model.SearchOutboundResourceGroup{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}

	if permission.Rbac {
		list, endList, err = api.app.GetOutboundResourceGroupPageByGroups(session.Domain(in.DomainId), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetOutboundResourceGroupPage(session.Domain(in.DomainId), req)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.OutboundResourceViewGroup, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.OutboundResourceViewGroup{
			Id:          v.Id,
			Name:        v.Name,
			Strategy:    v.Strategy,
			Description: v.Description,
			Communication: &engine.Lookup{
				Id:   int64(v.Communication.Id),
				Name: v.Communication.Name,
			},
		})
	}
	return &engine.ListOutboundResourceGroup{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *outboundResourceGroup) ReadOutboundResourceGroup(ctx context.Context, in *engine.ReadOutboundResourceGroupRequest) (*engine.OutboundResourceGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var group *model.OutboundResourceGroup

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	group, err = api.app.GetOutboundResourceGroup(session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceGroup(group), nil
}

func (api *outboundResourceGroup) UpdateOutboundResourceGroup(ctx context.Context, in *engine.UpdateOutboundResourceGroupRequest) (*engine.OutboundResourceGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	var group *model.OutboundResourceGroup

	group, err = api.app.UpdateOutboundResourceGroup(&model.OutboundResourceGroup{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.GetName(),
		Strategy:    in.GetStrategy(),
		Description: in.GetDescription(),
		Communication: model.Lookup{
			Id: int(in.GetCommunication().GetId()),
		},
		Time: toModelOutboundResourceGroupTime(in.Time),
	})

	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceGroup(group), nil
}

func (api *outboundResourceGroup) DeleteOutboundResourceGroup(ctx context.Context, in *engine.DeleteOutboundResourceGroupRequest) (*engine.OutboundResourceGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var group *model.OutboundResourceGroup
	group, err = api.app.RemoveOutboundResourceGroup(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceGroup(group), nil
}

func (api *outboundResourceGroup) CreateOutboundResourceInGroup(ctx context.Context, in *engine.CreateOutboundResourceInGroupRequest) (*engine.OutboundResourceInGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetGroupId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetGroupId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var res *model.OutboundResourceInGroup
	res, err = api.app.CreateOutboundResourceInGroup(session.Domain(in.GetDomainId()), in.GetResource().GetId(), in.GetGroupId())
	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceInGroup(res), nil
}

func (api *outboundResourceGroup) SearchOutboundResourceInGroup(ctx context.Context, in *engine.SearchOutboundResourceInGroupRequest) (*engine.ListOutboundResourceInGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetGroupId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetGroupId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.OutboundResourceInGroup
	list, err = api.app.GetOutboundResourceInGroupPage(session.Domain(in.DomainId), in.GetGroupId(), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.OutboundResourceInGroup, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineOutboundResourceInGroup(v))
	}
	return &engine.ListOutboundResourceInGroup{
		Items: items,
	}, nil
}

func (api *outboundResourceGroup) ReadOutboundResourceInGroup(ctx context.Context, in *engine.ReadOutboundResourceInGroupRequest) (*engine.OutboundResourceInGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetGroupId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetGroupId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var res *model.OutboundResourceInGroup
	res, err = api.app.GetOutboundResourceInGroup(session.Domain(in.DomainId), in.GetGroupId(), in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceInGroup(res), nil
}

func (api *outboundResourceGroup) UpdateOutboundResourceInGroup(ctx context.Context, in *engine.UpdateOutboundResourceInGroupRequest) (*engine.OutboundResourceInGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetGroupId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetGroupId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var res *model.OutboundResourceInGroup
	res, err = api.app.UpdateOutboundResourceInGroup(session.Domain(in.GetDomainId()), &model.OutboundResourceInGroup{
		Id:      in.GetId(),
		GroupId: in.GetGroupId(),
		Resource: model.Lookup{
			Id: int(in.GetResource().GetId()),
		},
	})

	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceInGroup(res), nil
}

func (api *outboundResourceGroup) DeleteOutboundResourceInGroup(ctx context.Context, in *engine.DeleteOutboundResourceInGroupRequest) (*engine.OutboundResourceInGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.OutboundResourceGroupCheckAccess(session.Domain(in.GetDomainId()), in.GetGroupId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetGroupId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var res *model.OutboundResourceInGroup
	res, err = api.app.RemoveOutboundResourceInGroup(session.Domain(in.GetDomainId()), in.GetGroupId(), in.GetId())

	if err != nil {
		return nil, err
	}

	return toEngineOutboundResourceInGroup(res), nil
}

func toEngineOutboundResourceGroup(src *model.OutboundResourceGroup) *engine.OutboundResourceGroup {
	return &engine.OutboundResourceGroup{
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
		Name:        src.Name,
		Strategy:    src.Strategy,
		Description: src.Description,
		Communication: &engine.Lookup{
			Id:   int64(src.Communication.Id),
			Name: src.Communication.Name,
		},
		Time: toEngineOutboundResourceTimeRange(src.Time),
	}
}

func toEngineOutboundResourceInGroup(src *model.OutboundResourceInGroup) *engine.OutboundResourceInGroup {
	return &engine.OutboundResourceInGroup{
		Id:      src.Id,
		GroupId: src.GroupId,
		Resource: &engine.Lookup{
			Id:   int64(src.Resource.Id),
			Name: src.Resource.Name,
		},
	}
}

func toEngineOutboundResourceTimeRange(src []model.OutboundResourceGroupTime) []*engine.OutboundResourceTimeRange {
	res := make([]*engine.OutboundResourceTimeRange, 0, len(src))

	for _, v := range src {
		res = append(res, &engine.OutboundResourceTimeRange{
			StartTimeOfDay: int32(v.StartTimeOfDay),
			EndTimeOfDay:   int32(v.EndTimeOfDay),
		})
	}

	return res
}

func toModelOutboundResourceGroupTime(src []*engine.OutboundResourceTimeRange) []model.OutboundResourceGroupTime {
	res := make([]model.OutboundResourceGroupTime, 0, len(src))

	for _, v := range src {
		res = append(res, model.OutboundResourceGroupTime{
			StartTimeOfDay: int16(v.StartTimeOfDay),
			EndTimeOfDay:   int16(v.EndTimeOfDay),
		})
	}

	return res
}
