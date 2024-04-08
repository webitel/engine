package grpc_api

import (
	"context"

	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

type outboundResource struct {
	app *app.App
	gogrpc.UnsafeOutboundResourceServiceServer
}

func NewOutboundResourceApi(app *app.App) *outboundResource {
	return &outboundResource{app: app}
}

func (api *outboundResource) CreateOutboundResource(ctx context.Context, in *engine.CreateOutboundResourceRequest) (*engine.OutboundResource, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	resource := &model.OutboundCallResource{
		DomainRecord: model.DomainRecord{
			DomainId:  session.Domain(0),
			CreatedAt: model.GetMillis(),
			CreatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Limit:                 int(in.Limit),
		Enabled:               in.Enabled,
		RPS:                   int(in.Rps),
		Reserve:               in.Reserve,
		Variables:             in.Variables,
		Number:                in.Number,
		MaxSuccessivelyErrors: int(in.MaxSuccessivelyErrors),
		Name:                  in.Name,
		ErrorIds:              in.ErrorIds,
		Description:           GetStringPointer(in.Description),
		Patterns:              in.Patterns,
		FailureDialDelay:      in.FailureDialDelay,
		Parameters: model.OutboundResourceParameters{
			CidType:          in.GetParameters().GetCidType(),
			IgnoreEarlyMedia: in.GetParameters().GetIgnoreEarlyMedia(),
		},
	}

	if in.Gateway != nil {
		resource.Gateway = &model.Lookup{
			Id: int(in.GetGateway().GetId()),
		}
	}

	if err = resource.IsValid(); err != nil {
		return nil, err
	}
	resource, err = api.app.CreateOutboundResource(ctx, resource)
	if err != nil {
		return nil, err
	}
	res := transformOutboundResource(resource)
	api.app.AuditCreate(ctx, session, model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE, res.Id, res)

	return res, nil

}

func (api *outboundResource) SearchOutboundResource(ctx context.Context, in *engine.SearchOutboundResourceRequest) (*engine.ListOutboundResource, error) {

	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.OutboundCallResource
	var endList bool
	req := &model.SearchOutboundCallResource{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.GetOutboundResourcePageByGroups(ctx, session.Domain(0), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetOutboundResourcePage(ctx, session.Domain(0), req)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.OutboundResource, 0, len(list))
	for _, v := range list {
		items = append(items, transformOutboundResource(v))
	}
	return &engine.ListOutboundResource{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *outboundResource) ReadOutboundResource(ctx context.Context, in *engine.ReadOutboundResourceRequest) (*engine.OutboundResource, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var resource *model.OutboundCallResource

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	resource, err = api.app.GetOutboundResource(ctx, session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return transformOutboundResource(resource), nil
}

func (api *outboundResource) UpdateOutboundResource(ctx context.Context, in *engine.UpdateOutboundResourceRequest) (*engine.OutboundResource, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(0), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	resource := &model.OutboundCallResource{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(0),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Limit:                 int(in.Limit),
		Enabled:               in.Enabled,
		RPS:                   int(in.Rps),
		Reserve:               in.Reserve,
		Variables:             in.Variables,
		Number:                in.Number,
		MaxSuccessivelyErrors: int(in.MaxSuccessivelyErrors),
		Name:                  in.Name,
		ErrorIds:              in.ErrorIds,
		Description:           GetStringPointer(in.Description),
		Patterns:              in.Patterns,
		FailureDialDelay:      in.FailureDialDelay,
		Parameters: model.OutboundResourceParameters{
			CidType:          in.GetParameters().GetCidType(),
			IgnoreEarlyMedia: in.GetParameters().GetIgnoreEarlyMedia(),
		},
	}

	if in.Gateway != nil {
		resource.Gateway = &model.Lookup{
			Id: int(in.GetGateway().GetId()),
		}
	}

	resource, err = api.app.UpdateOutboundResource(ctx, resource)

	if err != nil {
		return nil, err
	}

	res := transformOutboundResource(resource)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE, res.Id, res)

	return res, nil
}

func (api *outboundResource) PatchOutboundResource(ctx context.Context, in *engine.PatchOutboundResourceRequest) (*engine.OutboundResource, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(0), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	patch := &model.OutboundCallResourcePath{}
	var resource *model.OutboundCallResource

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "limit":
			patch.Limit = model.NewInt(int(in.Limit))
		case "rps":
			patch.RPS = model.NewInt(int(in.Rps))
		case "max_successively_errors":
			patch.MaxSuccessivelyErrors = model.NewInt(int(in.MaxSuccessivelyErrors))
		case "enabled":
			patch.Enabled = model.NewBool(in.Enabled)
		case "reserve":
			patch.Reserve = model.NewBool(in.Reserve)
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "failure_dial_delay":
			patch.FailureDialDelay = &in.FailureDialDelay
		}
	}

	resource, err = api.app.PatchOutboundResource(ctx, session.Domain(0), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	res := transformOutboundResource(resource)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE, res.Id, res)

	return res, nil
}

func (api *outboundResource) DeleteOutboundResource(ctx context.Context, in *engine.DeleteOutboundResourceRequest) (*engine.OutboundResource, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var resource *model.OutboundCallResource
	resource, err = api.app.RemoveOutboundResource(ctx, session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	res := transformOutboundResource(resource)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE, res.Id, res)

	return res, nil
}

func (api *outboundResource) CreateOutboundResourceDisplay(ctx context.Context, in *engine.CreateOutboundResourceDisplayRequest) (*engine.ResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	display := &model.ResourceDisplay{
		Display:    in.GetDisplay(),
		ResourceId: in.GetResourceId(),
	}

	if err = display.IsValid(); err != nil {
		return nil, err
	}

	display, err = api.app.CreateOutboundResourceDisplay(ctx, display)

	if err != nil {
		return nil, err
	}

	return toEngineResourceDisplay(display), nil
}

func (api *outboundResource) CreateOutboundResourceDisplayBulk(ctx context.Context, in *engine.CreateOutboundResourceDisplayBulkRequest) (*engine.ListResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.DomainId, in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var displays []*model.ResourceDisplay
	for _, disp := range in.Items {
		display := &model.ResourceDisplay{Display: disp.Display, ResourceId: in.ResourceId}
		if err = display.IsValid(); err != nil {
			return nil, err
		}
		displays = append(displays, display)
	}

	ids, err := api.app.CreateOutboundResourceDisplays(ctx, in.ResourceId, displays)

	if err != nil {
		return nil, err
	}

	return &engine.ListResourceDisplay{Id: ids}, nil
}

func (api *outboundResource) SearchOutboundResourceDisplay(ctx context.Context, in *engine.SearchOutboundResourceDisplayRequest) (*engine.ListOutboundResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(0), in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.ResourceDisplay
	var endList bool
	req := &model.SearchResourceDisplay{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.app.GetOutboundResourceDisplayPage(ctx, session.Domain(0), in.GetResourceId(), req)

	items := make([]*engine.ResourceDisplay, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineResourceDisplay(v))
	}
	return &engine.ListOutboundResourceDisplay{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *outboundResource) ReadOutboundResourceDisplay(ctx context.Context, in *engine.ReadOutboundResourceDisplayRequest) (*engine.ResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var display *model.ResourceDisplay
	display, err = api.app.GetOutboundResourceDisplay(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), in.GetId())

	if err != nil {
		return nil, err
	} else {
		return toEngineResourceDisplay(display), nil
	}
}

func (api *outboundResource) UpdateOutboundResourceDisplay(ctx context.Context, in *engine.UpdateOutboundResourceDisplayRequest) (*engine.ResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	display := &model.ResourceDisplay{
		Id:         in.GetId(),
		Display:    in.GetDisplay(),
		ResourceId: in.GetResourceId(),
	}

	if err = display.IsValid(); err != nil {
		return nil, err
	}

	display, err = api.app.UpdateOutboundResourceDisplay(ctx, session.Domain(in.GetDomainId()), display)

	if err != nil {
		return nil, err
	}

	return toEngineResourceDisplay(display), nil
}

func (api *outboundResource) DeleteOutboundResourceDisplay(ctx context.Context, in *engine.DeleteOutboundResourceDisplayRequest) (*engine.ResourceDisplay, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var display *model.ResourceDisplay
	display, err = api.app.RemoveOutboundResourceDisplay(ctx, session.Domain(in.GetDomainId()), in.GetResourceId(), in.GetId())

	if err != nil {
		return nil, err
	} else {
		return toEngineResourceDisplay(display), nil
	}

}

func (api *outboundResource) DeleteOutboundResourceDisplays(ctx context.Context, in *engine.DeleteOutboundResourceDisplaysRequest) (*engine.EmptyResponse, error) {
	empty := &engine.EmptyResponse{}
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, nil
	}
	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.OutboundResourceCheckAccess(ctx, session.DomainId, in.GetResourceId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetResourceId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	err = api.app.RemoveOutboundResourceDisplays(ctx, in.GetResourceId(), in.GetItems())

	if err != nil {
		return nil, err
	} else {
		return empty, nil
	}
}

func toEngineResourceDisplay(src *model.ResourceDisplay) *engine.ResourceDisplay {
	return &engine.ResourceDisplay{
		Id:      src.Id,
		Display: src.Display,
	}
}

func transformOutboundResource(src *model.OutboundCallResource) *engine.OutboundResource {
	res := &engine.OutboundResource{
		Id:                    src.Id,
		DomainId:              src.DomainId,
		CreatedAt:             src.CreatedAt,
		CreatedBy:             GetProtoLookup(src.CreatedBy),
		UpdatedAt:             src.UpdatedAt,
		UpdatedBy:             GetProtoLookup(src.UpdatedBy),
		Limit:                 int32(src.Limit),
		Enabled:               src.Enabled,
		Rps:                   int32(src.RPS),
		Reserve:               src.Reserve,
		Number:                src.Number,
		MaxSuccessivelyErrors: int32(src.MaxSuccessivelyErrors),
		Name:                  src.Name,
		Variables:             src.Variables,
		ErrorIds:              src.ErrorIds,
		LastErrorId:           src.LastError(),
		SuccessivelyErrors:    int32(src.SuccessivelyErrors),
		LastErrorAt:           model.TimeToInt64(src.LastErrorAt),
		Patterns:              src.Patterns,
		FailureDialDelay:      src.FailureDialDelay,
		Parameters: &engine.OutboundResourceParameters{
			CidType:          src.Parameters.CidType,
			IgnoreEarlyMedia: src.Parameters.IgnoreEarlyMedia,
		},
	}

	if src.Gateway != nil {
		res.Gateway = &engine.Lookup{
			Id:   int64(src.Gateway.Id),
			Name: src.Gateway.Name,
		}
	}

	if src.Description != nil {
		res.Description = *src.Description
	}

	return res
}
