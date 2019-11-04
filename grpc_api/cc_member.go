package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type member struct {
	app *app.App
}

func NewMemberApi(app *app.App) *member {
	return &member{app: app}
}

func (api *member) CreateMember(ctx context.Context, in *engine.CreateMemberRequest) (*engine.Member, error) {
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
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	member := &model.Member{
		QueueId:   in.GetQueueId(),
		Priority:  int(in.GetPriority()),
		Name:      in.GetName(),
		Variables: in.GetVariables(),
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Communications: toModelMemberCommunications(in.GetCommunications()),
	}

	if in.Bucket != nil {
		member.Bucket = &model.Lookup{
			Id: int(in.GetBucket().GetId()),
		}
	}

	if in.GetExpireAt() != 0 {
		member.ExpireAt = model.NewInt64(in.GetExpireAt())
	}

	if in.Bucket != nil {
		member.Bucket = &model.Lookup{
			Id: int(in.GetBucket().GetId()),
		}
	}

	if err = member.IsValid(); err != nil {
		return nil, err
	}

	if member, err = api.app.CreateMember(member); err != nil {
		return nil, err
	}

	return toEngineMember(member), nil
}

func (api *member) CreateMemberBulk(ctx context.Context, in *engine.CreateMemberBulkRequest) (*engine.MemberBulkResponse, error) {
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
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	members := make([]*model.Member, 0, len(in.Items))
	for _, v := range in.Items {
		member := &model.Member{
			QueueId:   in.GetQueueId(),
			Priority:  int(v.GetPriority()),
			Name:      v.GetName(),
			Variables: v.GetVariables(),
			Timezone: model.Lookup{
				Id: int(v.GetTimezone().GetId()),
			},
			Communications: toModelMemberCommunications(v.GetCommunications()),
		}
		if v.GetExpireAt() != 0 {
			member.ExpireAt = model.NewInt64(v.GetExpireAt())
		}

		if v.Bucket != nil {
			member.Bucket = &model.Lookup{
				Id: int(v.GetBucket().GetId()),
			}
		}

		if err = member.IsValid(); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	err = api.app.BulkCreateMember(session.Domain(in.GetDomainId()), in.GetQueueId(), members)
	if err != nil {
		return nil, err
	}

	return &engine.MemberBulkResponse{}, nil
}

func (api *member) ReadMember(ctx context.Context, in *engine.ReadMemberRequest) (*engine.Member, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.Member
	out, err = api.app.GetMember(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineMember(out), nil
}

func (api *member) SearchMember(ctx context.Context, in *engine.SearchMemberRequest) (*engine.ListMember, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.Member

	if list, err = api.app.GetMemberPage(session.Domain(in.GetDomainId()), in.GetQueueId(), int(in.GetPage()), int(in.GetSize())); err != nil {
		return nil, err
	}

	items := make([]*engine.Member, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineMember(v))
	}

	return &engine.ListMember{
		Items: items,
	}, nil
}

func (api *member) UpdateMember(ctx context.Context, in *engine.UpdateMemberRequest) (*engine.Member, error) {
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
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	member := &model.Member{
		Id:        in.GetId(),
		QueueId:   in.GetQueueId(),
		Priority:  int(in.GetPriority()),
		ExpireAt:  nil,
		Bucket:    nil,
		Name:      in.GetName(),
		Variables: in.GetVariables(),
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Communications: toModelMemberCommunications(in.GetCommunications()),
	}

	if in.ExpireAt != 0 {
		member.ExpireAt = model.NewInt64(in.ExpireAt)
	} else {
		member.ExpireAt = nil
	}

	if in.Bucket != nil {
		member.Bucket = &model.Lookup{
			Id: int(in.GetBucket().GetId()),
		}
	}

	if err = member.IsValid(); err != nil {
		return nil, err
	}

	if member, err = api.app.UpdateMember(session.Domain(in.GetDomainId()), member); err != nil {
		return nil, err
	} else {
		return toEngineMember(member), nil
	}
}

func (api *member) DeleteMember(ctx context.Context, in *engine.DeleteMemberRequest) (*engine.Member, error) {
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
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var member *model.Member
	member, err = api.app.RemoveMember(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineMember(member), nil
}

func toEngineMember(src *model.Member) *engine.Member {
	res := &engine.Member{
		Id:        src.Id,
		QueueId:   src.QueueId,
		Priority:  int32(src.Priority),
		ExpireAt:  src.GetExpireAt(),
		Variables: src.Variables,
		Name:      src.Name,
		Timezone: &engine.Lookup{
			Id:   int64(src.Timezone.Id),
			Name: src.Name,
		},
		Communications: toEngineMemberCommunications(src.Communications),
		LastActivityAt: src.LastActivityAt,
		Attempts:       int32(src.Attempts),
	}

	if src.Bucket != nil {
		res.Bucket = &engine.Lookup{
			Id:   int64(src.Bucket.Id),
			Name: src.Bucket.Name,
		}
	}

	if src.StopCause != nil {
		res.StopCause = *src.StopCause
	}

	return res
}

func toEngineMemberCommunications(src []model.MemberCommunication) []*engine.MemberCommunication {
	res := make([]*engine.MemberCommunication, 0, len(src))

	for _, v := range src {
		res = append(res, &engine.MemberCommunication{
			Priority:       int32(v.Priority),
			Destination:    v.Destination,
			State:          int32(v.State),
			Description:    v.Description,
			LastActivityAt: v.LastActivityAt,
			Attempts:       int32(v.Attempts),
			LastCause:      v.LastCause,
		})
	}

	return res
}

func toModelMemberCommunications(src []*engine.MemberCommunicationCreateRequest) []model.MemberCommunication {
	res := make([]model.MemberCommunication, 0, len(src))

	for _, v := range src {
		res = append(res, model.MemberCommunication{
			Priority:    int(v.GetPriority()),
			Destination: v.GetDestination(),
			Description: v.GetDescription(),
		})
	}

	return res
}
