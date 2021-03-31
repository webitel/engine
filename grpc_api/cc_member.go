package grpc_api

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"net/http"
	"strings"
)

type member struct {
	*API
}

func NewMemberApi(api *API) *member {
	return &member{api}
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
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
		MinOfferingAt:  model.Int64ToTime(in.MinOfferingAt),
		ExpireAt:       model.Int64ToTime(in.GetExpireAt()),

		Bucket: GetLookup(in.Bucket),
		Agent:  GetLookup(in.Agent),
		Skill:  GetLookup(in.Skill),
	}

	if err = member.IsValid(); err != nil {
		return nil, err
	}

	if member, err = api.app.CreateMember(session.DomainId, member); err != nil {
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
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
			MinOfferingAt:  model.Int64ToTime(v.MinOfferingAt),
			ExpireAt:       model.Int64ToTime(v.GetExpireAt()),
			Bucket:         GetLookup(v.GetBucket()),
			Agent:          GetLookup(v.Agent),
			Skill:          GetLookup(v.Skill),
		}

		if err = member.IsValid(); err != nil {
			return nil, err
		}

		members = append(members, member)
	}
	var inserted []int64

	inserted, err = api.app.BulkCreateMember(session.Domain(in.GetDomainId()), in.GetQueueId(), members)
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.Member
	out, err = api.app.GetMember(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.Member
	var endList bool
	req := &model.SearchMemberRequest{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}

	if list, endList, err = api.app.GetMemberPage(session.Domain(in.GetDomainId()), in.GetQueueId(), req); err != nil {
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	member := &model.Member{
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
		member.StopCause = &in.StopCause
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var member *model.Member
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
			patch.Bucket = GetLookup(in.Bucket)
		case "communications":
			patch.Communications = toModelMemberCommunications(in.GetCommunications())
		case "stop_cause":
			patch.StopCause = model.NewString(in.StopCause)
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

	member, err = api.app.PatchMember(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return toEngineMember(member), nil

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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var member *model.Member
	member, err = api.app.RemoveMember(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineMember(member), nil
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

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var list []*model.Member

	list, err = api.app.RemoveMultiMembers(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetIds(), in.GetBucketId(), in.GetStopCause())

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

func (api *member) SearchMemberAttempts(ctx context.Context, in *engine.SearchMemberAttemptsRequest) (*engine.ListMemberAttempt, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.MemberAttempt
	if list, err = api.app.GetMemberAttempts(in.GetMemberId()); err != nil {
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

	if in.GetJoinedAt() == nil {
		return nil, model.NewAppError("GRPC.SearchAttempts", "grpc.member.search_attempt", nil, "filter joined_at is required", http.StatusBadRequest)
	}

	var list []*model.Attempt
	var endList bool
	req := &model.SearchAttempts{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Fields:   in.GetFields(),
			Sort:     in.GetSort(),
		},
		JoinedAt: model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		},
		Ids:       in.GetId(),
		MemberIds: in.GetMemberId(),
		QueueIds:  in.GetQueueId(),
		BucketIds: in.GetBucketId(),
		AgentIds:  in.GetAgentId(),
	}

	if in.GetResult() != "" {
		req.Result = &in.Result
	}

	if list, endList, err = api.app.SearchAttempts(session.Domain(in.GetDomainId()), req); err != nil {
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
		return nil, model.NewAppError("GRPC.SearchAttempts", "grpc.member.search_attempt", nil, "filter joined_at is required", http.StatusBadRequest)
	}

	var list []*model.AttemptHistory
	var endList bool
	req := &model.SearchAttempts{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Fields:   in.GetFields(),
			Sort:     in.GetSort(),
		},
		JoinedAt: model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		},
		Ids:       in.GetId(),
		MemberIds: in.GetMemberId(),
		QueueIds:  in.GetQueueId(),
		BucketIds: in.GetBucketId(),
		AgentIds:  in.GetAgentId(),
	}

	if in.GetResult() != "" {
		req.Result = &in.Result
	}

	if list, endList, err = api.app.SearchAttemptsHistory(session.Domain(in.GetDomainId()), req); err != nil {
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
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}

	if in.GetId() != 0 {
		req.Id = &in.Id
	}

	if in.GetQueueId() != 0 {
		req.QueueId = &in.QueueId
	}

	if in.GetDestination() != "" {
		req.Destination = &in.Destination
	}

	if in.GetBucketId() != 0 {
		req.BucketId = &in.BucketId
	}

	if list, endList, err = api.app.SearchMembers(session.Domain(in.GetDomainId()), req); err != nil {
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

	if in.MinOfferingAt > 0 {
		nextOffering = &in.MinOfferingAt
	}

	if in.ExpireAt > 0 {
		expire = &in.ExpireAt
	}

	err = api.ctrl.ReportingAttempt(session, in.AttemptId, in.Status, in.Description, nextOffering, expire, in.Variables, in.Display, in.AgentId)
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

	if in.MinOfferingAt > 0 {
		nextOffering = &in.MinOfferingAt
	}

	if in.ExpireAt > 0 {
		expire = &in.ExpireAt
	}

	err = api.ctrl.ReportingAttempt(session, in.AttemptId, in.Status, in.Description, nextOffering, expire, in.Variables, in.Display, in.AgentId)
	if err != nil {
		return nil, err
	}

	return &engine.AttemptResultResponse{
		Status: "success",
	}, nil
}

func toEngineMemberCommunications(src []model.MemberCommunication) []*engine.MemberCommunication {
	res := make([]*engine.MemberCommunication, 0, len(src))

	for _, v := range src {
		res = append(res, toEngineDestination(v))
	}

	return res
}

func toEngineDestination(v model.MemberCommunication) *engine.MemberCommunication {
	return &engine.MemberCommunication{
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
}

func toModelMemberCommunications(src []*engine.MemberCommunicationCreateRequest) []model.MemberCommunication {
	res := make([]model.MemberCommunication, 0, len(src))

	for _, v := range src {
		res = append(res, model.MemberCommunication{
			Priority:    int(v.GetPriority()),
			Destination: v.GetDestination(),
			Description: v.GetDescription(),
			Type: model.Lookup{
				Id: int(v.GetType().GetId()),
			},
			Resource: GetLookup(v.Resource),
			Display:  v.Display,
		})
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
		Queue:           GetProtoLookup(&src.Queue),
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
		Queue:        GetProtoLookup(&src.Queue),
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
	}

	if src.MemberCallId != nil {
		item.MemberCallId = *src.MemberCallId
	}

	if src.AgentCallId != nil {
		item.AgentCallId = *src.AgentCallId
	}

	return item
}
