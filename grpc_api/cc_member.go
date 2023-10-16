package grpc_api

import (
	"context"
	"strings"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type member struct {
	*API
	engine.UnsafeMemberServiceServer
}

func NewMemberApi(api *API) *member {
	return &member{API: api}
}

func (api *member) CreateMember(ctx context.Context, in *engine.CreateMemberRequest) (*engine.MemberInQueue, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	m := &model.Member{
		QueueId:   in.GetQueueId(),
		Priority:  int(in.GetPriority()),
		Name:      in.GetName(),
		Variables: in.GetVariables(),
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Communications: toModelMemberCommunications(in.GetCommunications()),
		MinOfferingAt:  model.Int64ToTime(in.MinOfferingAt),
		ExpireAt:       model.Int64ToTime(in.GetExpireAt()),

		Bucket: GetLookup(in.Bucket),
		Agent:  GetLookup(in.Agent),
		Skill:  GetLookup(in.Skill),
	}

	if err = m.IsValid(api.app.MaxMemberCommunications()); err != nil {
		return nil, err
	}

	if m, err = api.app.CreateMember(ctx, session.DomainId, m); err != nil {
		return nil, err
	}

	return toEngineMember(m), nil
}

func (api *member) CreateMemberBulk(ctx context.Context, in *engine.CreateMemberBulkRequest) (*engine.MemberBulkResponse, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(0), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	members := make([]*model.Member, 0, len(in.Items))
	for _, v := range in.Items {
		m := &model.Member{
			QueueId:   in.GetQueueId(),
			Priority:  int(v.GetPriority()),
			Name:      v.GetName(),
			Variables: v.GetVariables(),
			Timezone: model.Lookup{
				Id: int(v.GetTimezone().GetId()),
			},
			Communications: toModelMemberCommunications(v.GetCommunications()),
			MinOfferingAt:  model.Int64ToTime(v.MinOfferingAt),
			ExpireAt:       model.Int64ToTime(v.GetExpireAt()),
			Bucket:         GetLookup(v.GetBucket()),
			Agent:          GetLookup(v.Agent),
			Skill:          GetLookup(v.Skill),
		}

		if err = m.IsValid(api.app.MaxMemberCommunications()); err != nil {
			return nil, err
		}

		members = append(members, m)
	}
	var inserted []int64

	inserted, err = api.app.BulkCreateMember(ctx, session.Domain(0), in.GetQueueId(), in.GetFileName(), members)
	if err != nil {
		return nil, err
	}

	return &engine.MemberBulkResponse{
		Ids: inserted,
	}, nil
}

func (api *member) ReadMember(ctx context.Context, in *engine.ReadMemberRequest) (*engine.MemberInQueue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.Member
	out, err = api.app.GetMember(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineMember(out), nil
}

func (api *member) SearchMemberInQueue(ctx context.Context, in *engine.SearchMemberInQueueRequest) (*engine.ListMember, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(0), int64(in.GetQueueId()), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, int64(in.GetQueueId()), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.Member
	var endList bool
	req := &model.SearchMemberRequest{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		Ids:        in.GetId(),
		QueueId:    &in.QueueId,
		BucketIds:  in.GetBucketId(),
		StopCauses: in.GetStopCause(),
		AgentIds:   in.GetAgentId(),
	}

	if in.Destination != "" {
		req.Destination = &in.Destination
	}

	if in.Name != "" {
		req.Name = &in.Name
	}

	if in.GetPriority() != nil {
		req.Priority = &model.FilterBetween{
			From: in.GetPriority().GetFrom(),
			To:   in.GetPriority().GetTo(),
		}
	}

	if in.GetAttempts() != nil {
		req.Attempts = &model.FilterBetween{
			From: in.GetAttempts().GetFrom(),
			To:   in.GetAttempts().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}
	if in.GetOfferingAt() != nil {
		req.OfferingAt = &model.FilterBetween{
			From: in.GetOfferingAt().GetFrom(),
			To:   in.GetOfferingAt().GetTo(),
		}
	}

	if list, endList, err = api.app.SearchMembers(ctx, session.Domain(0), req); err != nil {
		return nil, err
	}

	items := make([]*engine.MemberInQueue, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineMember(v))
	}

	return &engine.ListMember{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *member) UpdateMember(ctx context.Context, in *engine.UpdateMemberRequest) (*engine.MemberInQueue, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	m := &model.Member{
		Id:        in.GetId(),
		QueueId:   in.GetQueueId(),
		Priority:  int(in.GetPriority()),
		Name:      in.GetName(),
		Variables: in.GetVariables(),
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Communications: toModelMemberCommunications(in.GetCommunications()),
		MinOfferingAt:  model.Int64ToTime(in.MinOfferingAt),
		ExpireAt:       model.Int64ToTime(in.ExpireAt),
		Bucket:         GetLookup(in.Bucket),
		Agent:          GetLookup(in.Agent),
		Skill:          GetLookup(in.Skill),
	}

	if in.StopCause != "" {
		m.StopCause = &in.StopCause
	}

	if err = m.IsValid(api.app.MaxMemberCommunications()); err != nil {
		return nil, err
	}

	if m, err = api.app.UpdateMember(ctx, session.Domain(in.GetDomainId()), m); err != nil {
		return nil, err
	} else {
		return toEngineMember(m), nil
	}
}

func (api *member) PatchMember(ctx context.Context, in *engine.PatchMemberRequest) (*engine.MemberInQueue, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var m *model.Member
	patch := &model.MemberPatch{}

	//TODO FIXME
	for _, v := range in.Fields {
		switch v {
		case "priority":
			patch.Priority = model.NewInt(int(in.Priority))
		case "expire_at":
			patch.ExpireAt = model.Int64ToTime(in.ExpireAt)
		case "min_offering_at":
			patch.MinOfferingAt = model.Int64ToTime(in.MinOfferingAt)
		case "name":
			patch.Name = model.NewString(in.Name)
		case "timezone.id":
			patch.Timezone = GetLookup(in.Timezone)
		case "bucket.id":
			//todo
			if in.Bucket != nil && in.Bucket.Id == 0 {
				patch.Bucket = &model.Lookup{
					Id: 0,
				}
			} else {
				patch.Bucket = GetLookup(in.Bucket)
			}
		case "communications":
			patch.Communications = toModelMemberCommunications(in.GetCommunications())
		case "stop_cause":
			patch.StopCause = model.NewString(in.StopCause)
		case "attempts":
			patch.Attempts = model.NewInt(int(in.Attempts))
		case "agent.id":
			//todo
			if in.Agent != nil && in.Agent.Id == 0 {
				patch.Agent = &model.Lookup{
					Id: 0,
				}
			} else {
				patch.Agent = GetLookup(in.Agent)
			}
		case "skill.id":
			//todo
			if in.Skill != nil && in.Skill.Id == 0 {
				patch.Skill = &model.Lookup{
					Id: 0,
				}
			} else {
				patch.Skill = GetLookup(in.Skill)
			}
		default:
			if patch.Variables == nil && strings.HasPrefix(v, "variables.") {
				patch.Variables = in.Variables
			}
		}
	}

	m, err = api.app.PatchMember(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return toEngineMember(m), nil

}

func (api *member) DeleteMember(ctx context.Context, in *engine.DeleteMemberRequest) (*engine.MemberInQueue, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var m *model.Member
	m, err = api.app.RemoveMember(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineMember(m), nil
}

func (api *member) DeleteMembers(ctx context.Context, in *engine.DeleteMembersRequest) (*engine.ListMember, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(0), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var list []*model.Member

	req := &model.MultiDeleteMembers{
		QueueId: in.QueueId,
		SearchMemberRequest: model.SearchMemberRequest{
			ListRequest: model.ListRequest{
				Q:       in.GetQ(),
				PerPage: int(in.GetSize()),
				Sort:    in.GetSort(),
			},
			Ids:        in.GetId(),
			QueueIds:   []int32{int32(in.GetQueueId())},
			BucketIds:  in.GetBucketId(),
			StopCauses: in.GetStopCause(),
			AgentIds:   in.GetAgentId(),
		},
		Numbers:   in.GetNumbers(),
		Variables: in.GetVariables(),
	}

	//todo deprecated
	if in.GetIds() != nil {
		req.Ids = in.GetIds()
	}

	if in.Destination != "" {
		req.Destination = &in.Destination
	}

	if in.Name != "" {
		req.Name = &in.Name
	}

	if in.GetPriority() != nil {
		req.Priority = &model.FilterBetween{
			From: in.GetPriority().GetFrom(),
			To:   in.GetPriority().GetTo(),
		}
	}

	if in.GetAttempts() != nil {
		req.Attempts = &model.FilterBetween{
			From: in.GetAttempts().GetFrom(),
			To:   in.GetAttempts().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}
	if in.GetOfferingAt() != nil {
		req.OfferingAt = &model.FilterBetween{
			From: in.GetOfferingAt().GetFrom(),
			To:   in.GetOfferingAt().GetTo(),
		}
	}

	list, err = api.app.RemoveMultiMembers(ctx, session.Domain(0), req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.MemberInQueue, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineMember(v))
	}

	return &engine.ListMember{
		Items: items,
	}, nil
}

func (api *member) ResetMembers(ctx context.Context, in *engine.ResetMembersRequest) (*engine.ResetMembersResponse, error) {
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
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(0), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	var cnt int64
	search := &model.ResetMembers{
		QueueId:   in.GetQueueId(),
		Ids:       in.GetIds(),
		Buckets:   in.GetBucketId(),
		Causes:    in.GetStopCause(),
		AgentIds:  in.GetAgentId(),
		Numbers:   in.GetNumbers(),
		Variables: in.GetVariables(),
	}

	if len(search.Ids) == 0 {
		search.Ids = in.GetId()
	}

	cnt, err = api.app.ResetMembers(ctx, session.Domain(0), search)

	if err != nil {
		return nil, err
	}

	return &engine.ResetMembersResponse{
		Count: cnt,
	}, nil
}

func (api *member) SearchMemberAttempts(ctx context.Context, in *engine.SearchMemberAttemptsRequest) (*engine.ListMemberAttempt, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.MemberAttempt
	if list, err = api.app.GetMemberAttempts(ctx, in.GetMemberId()); err != nil {
		return nil, err
	}

	items := make([]*engine.MemberAttempt, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineMemberAttempt(v))
	}

	return &engine.ListMemberAttempt{
		Items: items,
	}, nil
}

func (api *member) SearchAttempts(ctx context.Context, in *engine.SearchAttemptsRequest) (*engine.ListAttempt, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	//FIXME check queue PERMISSION

	var list []*model.Attempt
	var endList bool
	req := &model.SearchAttempts{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		Ids:       in.GetId(),
		MemberIds: in.GetMemberId(),
		QueueIds:  in.GetQueueId(),
		BucketIds: in.GetBucketId(),
		AgentIds:  in.GetAgentId(),
		Result:    in.GetResult(),
	}

	if in.JoinedAt != nil {
		req.JoinedAt = &model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		}
	}

	if in.LeavingAt != nil {
		req.LeavingAt = &model.FilterBetween{
			From: in.GetLeavingAt().GetFrom(),
			To:   in.GetLeavingAt().GetTo(),
		}
	}

	if in.OfferingAt != nil {
		req.OfferingAt = &model.FilterBetween{
			From: in.GetOfferingAt().GetFrom(),
			To:   in.GetOfferingAt().GetTo(),
		}
	}

	if in.Duration != nil {
		req.Duration = &model.FilterBetween{
			From: in.GetDuration().GetFrom(),
			To:   in.GetDuration().GetTo(),
		}
	}

	if list, endList, err = api.app.SearchAttempts(ctx, session.Domain(0), req); err != nil {
		return nil, err
	}

	items := make([]*engine.Attempt, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineAttempt(v))
	}

	return &engine.ListAttempt{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *member) SearchAttemptsHistory(ctx context.Context, in *engine.SearchAttemptsRequest) (*engine.ListHistoryAttempt, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	//FIXME check queue PERMISSION

	if in.GetJoinedAt() == nil {
		return nil, model.NewBadRequestError("grpc.member.search_attempt", "filter joined_at is required")
	}

	var list []*model.AttemptHistory
	var endList bool
	req := &model.SearchAttempts{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		Ids:       in.GetId(),
		MemberIds: in.GetMemberId(),
		QueueIds:  in.GetQueueId(),
		BucketIds: in.GetBucketId(),
		AgentIds:  in.GetAgentId(),
		Result:    in.GetResult(),
	}

	if in.JoinedAt != nil {
		req.JoinedAt = &model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		}
	}

	if in.LeavingAt != nil {
		req.LeavingAt = &model.FilterBetween{
			From: in.GetLeavingAt().GetFrom(),
			To:   in.GetLeavingAt().GetTo(),
		}
	}

	if in.OfferingAt != nil {
		req.OfferingAt = &model.FilterBetween{
			From: in.GetOfferingAt().GetFrom(),
			To:   in.GetOfferingAt().GetTo(),
		}
	}

	if in.Duration != nil {
		req.Duration = &model.FilterBetween{
			From: in.GetDuration().GetFrom(),
			To:   in.GetDuration().GetTo(),
		}
	}

	if list, endList, err = api.app.SearchAttemptsHistory(ctx, session.Domain(0), req); err != nil {
		return nil, err
	}

	items := make([]*engine.AttemptHistory, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineAttemptHistory(v))
	}

	return &engine.ListHistoryAttempt{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *member) SearchMembers(ctx context.Context, in *engine.SearchMembersRequest) (*engine.ListMember, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	//FIXME check queue PERMISSION

	var list []*model.Member
	var endList bool
	req := &model.SearchMemberRequest{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		Ids:        in.GetId(),
		QueueIds:   in.GetQueueId(),
		BucketIds:  in.GetBucketId(),
		StopCauses: in.GetStopCause(),
		AgentIds:   in.GetAgentId(),
	}

	if in.Destination != "" {
		req.Destination = &in.Destination
	}

	if in.Name != "" {
		req.Name = &in.Name
	}

	if in.GetPriority() != nil {
		req.Priority = &model.FilterBetween{
			From: in.GetPriority().GetFrom(),
			To:   in.GetPriority().GetTo(),
		}
	}

	if in.GetAttempts() != nil {
		req.Attempts = &model.FilterBetween{
			From: in.GetAttempts().GetFrom(),
			To:   in.GetAttempts().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}
	if in.GetOfferingAt() != nil {
		req.OfferingAt = &model.FilterBetween{
			From: in.GetOfferingAt().GetFrom(),
			To:   in.GetOfferingAt().GetTo(),
		}
	}

	if list, endList, err = api.app.SearchMembers(ctx, session.Domain(0), req); err != nil {
		return nil, err
	}

	items := make([]*engine.MemberInQueue, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineMember(v))
	}

	return &engine.ListMember{
		Next:  !endList,
		Items: items,
	}, nil
}

func toEngineMemberAttempt(src *model.MemberAttempt) *engine.MemberAttempt {
	res := &engine.MemberAttempt{
		Id:          src.Id,
		CreatedAt:   src.CreatedAt,
		Destination: src.Destination,
		Weight:      int32(src.Weight),
		OriginateAt: src.OriginateAt,
		AnsweredAt:  src.AnsweredAt,
		BridgedAt:   src.BridgedAt,
		HangupAt:    src.HangupAt,
		Resource: &engine.Lookup{
			Id:   int64(src.Resource.Id),
			Name: src.Resource.Name,
		},
		Logs:   UnmarshalJsonpb(src.Logs),
		Active: src.Active,
	}

	if src.LegAId != nil {
		res.LegAId = *src.LegAId
	}
	if src.LegBId != nil {
		res.LegBId = *src.LegBId
	}

	if src.Node != nil {
		res.Node = *src.Node
	}

	if src.Result != nil {
		res.Result = *src.Result
	}

	if src.Agent != nil {
		res.Agent = &engine.Lookup{
			Id:   int64(src.Agent.Id),
			Name: src.Agent.Name,
		}
	}
	if src.Bucket != nil {
		res.Bucket = &engine.Lookup{
			Id:   int64(src.Bucket.Id),
			Name: src.Bucket.Name,
		}
	}

	if src.Attempts != nil {
		res.Attempts = *src.Attempts
	}

	return res
}

func toEngineMember(src *model.Member) *engine.MemberInQueue {
	res := &engine.MemberInQueue{
		Id:        src.Id,
		CreatedAt: model.TimeToInt64(&src.CreatedAt),
		Queue:     GetProtoLookup(&src.Queue),
		Priority:  int32(src.Priority),
		ExpireAt:  model.TimeToInt64(src.ExpireAt),
		Variables: src.Variables,
		Name:      src.Name,
		Timezone: &engine.Lookup{
			Id:   int64(src.Timezone.Id),
			Name: src.Timezone.Name,
		},
		Communications: toEngineMemberCommunications(src.Communications),
		LastActivityAt: src.LastActivityAt,
		Attempts:       int32(src.Attempts),
		MinOfferingAt:  model.TimeToInt64(src.MinOfferingAt),
		Reserved:       src.Reserved,
		Agent:          GetProtoLookup(src.Agent),
		Skill:          GetProtoLookup(src.Skill),
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

func (api *member) CreateAttempt(ctx context.Context, in *engine.CreateAttemptRequest) (*engine.CreateAttemptResponse, error) {
	//TODO validate && proxy cc
	return nil, nil
}

func (api *member) AttemptResult(ctx context.Context, in *engine.AttemptResultRequest) (*engine.AttemptResultResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var nextOffering *int64
	var expire *int64
	var waitBetweenRetries *int32

	if in.MinOfferingAt > 0 {
		nextOffering = &in.MinOfferingAt
	}

	if in.ExpireAt > 0 {
		expire = &in.ExpireAt
	}

	if in.WaitBetweenRetries > 0 {
		waitBetweenRetries = &in.WaitBetweenRetries
	}

	err = api.ctrl.ReportingAttempt(session, in.AttemptId, in.Status, in.Description, nextOffering, expire, in.Variables,
		in.Display, in.AgentId, in.ExcludeCurrentCommunication, waitBetweenRetries)
	if err != nil {
		return nil, err
	}

	return &engine.AttemptResultResponse{
		Status: "success",
	}, nil
}

func (api *member) AttemptCallback(ctx context.Context, in *engine.AttemptCallbackRequest) (*engine.AttemptResultResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var nextOffering *int64
	var expire *int64
	var waitBetweenRetries *int32

	if in.MinOfferingAt > 0 {
		nextOffering = &in.MinOfferingAt
	}

	if in.ExpireAt > 0 {
		expire = &in.ExpireAt
	}

	if in.WaitBetweenRetries > 0 {
		waitBetweenRetries = &in.WaitBetweenRetries
	}

	err = api.ctrl.ReportingAttempt(session, in.AttemptId, in.Status, in.Description, nextOffering, expire, in.Variables,
		in.Display, in.AgentId, in.ExcludeCurrentCommunication, waitBetweenRetries)
	if err != nil {
		return nil, err
	}

	return &engine.AttemptResultResponse{
		Status: "success",
	}, nil
}

func (api *member) AttemptsRenewalResult(ctx context.Context, in *engine.AttemptRenewalResultRequest) (*engine.AttemptRenewalResultResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = api.ctrl.RenewalAttempt(session, in.AttemptId, in.Renewal)
	if err != nil {
		return nil, err
	}

	return &engine.AttemptRenewalResultResponse{}, nil
}

func toEngineMemberCommunications(src []model.MemberCommunication) []*engine.MemberCommunication {
	res := make([]*engine.MemberCommunication, 0, len(src))

	for _, v := range src {
		res = append(res, toEngineDestination(v))
	}

	return res
}

func toEngineDestination(v model.MemberCommunication) *engine.MemberCommunication {
	c := &engine.MemberCommunication{
		Id:             v.Id,
		Priority:       int32(v.Priority),
		Destination:    v.Destination,
		State:          int32(v.State),
		Description:    v.Description,
		LastActivityAt: v.LastActivityAt,
		Attempts:       int32(v.Attempts),
		LastCause:      v.LastCause,
		Type: &engine.Lookup{
			Id:   int64(v.Type.Id),
			Name: v.Type.Name,
		},
		Resource: GetProtoLookup(v.Resource),
		Display:  v.Display,
	}
	if v.StopAt != nil {
		c.StopAt = *v.StopAt
	}
	if v.Dtmf != nil {
		c.Dtmf = *v.Dtmf
	}
	return c
}

func toModelMemberCommunications(src []*engine.MemberCommunicationCreateRequest) []model.MemberCommunication {
	res := make([]model.MemberCommunication, 0, len(src))

	for _, v := range src {
		c := model.MemberCommunication{
			Priority:    int(v.GetPriority()),
			Destination: strings.Trim(v.GetDestination(), " "),
			Description: v.GetDescription(),
			Type: model.Lookup{
				Id: int(v.GetType().GetId()),
			},
			Resource: GetLookup(v.Resource),
			Display:  v.Display,
		}

		if v.Dtmf != "" {
			c.Dtmf = model.NewString(v.Dtmf)
		}

		if v.GetStopAt() != 0 {
			c.StopAt = model.NewInt64(v.GetStopAt())
		}
		res = append(res, c)
	}

	return res
}

func toEngineAttempt(src *model.Attempt) *engine.Attempt {
	item := &engine.Attempt{
		Id:              src.Id,
		State:           src.State,
		LastStateChange: src.LastStateChange,
		JoinedAt:        src.JoinedAt,
		OfferingAt:      src.OfferingAt,
		BridgedAt:       src.BridgedAt,
		ReportingAt:     src.ReportingAt,
		Timeout:         src.Timeout,
		LeavingAt:       src.LeavingAt,
		Channel:         src.Channel,
		Queue:           GetProtoLookup(src.Queue),
		Member:          GetProtoLookup(src.Member),
		MemberCallId:    "",
		Variables:       src.Variables,
		Agent:           GetProtoLookup(src.Agent),
		AgentCallId:     "",
		Position:        int32(src.Position),
		Resource:        GetProtoLookup(src.Resource),
		Bucket:          GetProtoLookup(src.Bucket),
		List:            GetProtoLookup(src.List),
		Display:         src.Display,
		Destination:     toEngineDestination(src.Destination),
		Result:          "",
	}

	if src.MemberCallId != nil {
		item.MemberCallId = *src.MemberCallId
	}

	if src.AgentCallId != nil {
		item.AgentCallId = *src.AgentCallId
	}

	if src.Result != nil {
		item.Result = *src.Result
	}

	if src.Attempts != nil {
		item.Attempts = *src.Attempts
	}

	return item
}

func toEngineAttemptHistory(src *model.AttemptHistory) *engine.AttemptHistory {
	item := &engine.AttemptHistory{
		Id:           src.Id,
		JoinedAt:     model.TimeToInt64(src.JoinedAt),
		OfferingAt:   model.TimeToInt64(src.OfferingAt),
		BridgedAt:    model.TimeToInt64(src.BridgedAt),
		ReportingAt:  model.TimeToInt64(src.ReportingAt),
		LeavingAt:    model.TimeToInt64(src.LeavingAt),
		Channel:      src.Channel,
		Queue:        GetProtoLookup(src.Queue),
		Member:       GetProtoLookup(src.Member),
		MemberCallId: "",
		Variables:    src.Variables,
		Agent:        GetProtoLookup(src.Agent),
		AgentCallId:  "",
		Position:     int32(src.Position),
		Resource:     GetProtoLookup(src.Resource),
		Bucket:       GetProtoLookup(src.Bucket),
		List:         GetProtoLookup(src.List),
		Display:      src.Display,
		Destination:  toEngineDestination(src.Destination),
		Result:       src.Result,
		AmdResult:    defaultString(src.AmdResult),
	}

	if src.MemberCallId != nil {
		item.MemberCallId = *src.MemberCallId
	}

	if src.AgentCallId != nil {
		item.AgentCallId = *src.AgentCallId
	}

	if src.Attempts != nil {
		item.Attempts = *src.Attempts
	}

	return item
}
