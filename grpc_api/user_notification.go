package grpc_api

import (
	"context"

	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type userNotification struct {
	*API
	engine.UnsafeUserNotificationServiceServer
}

func NewUserNotificationApi(api *API) *userNotification {
	return &userNotification{API: api}
}

func (api *userNotification) SearchUserNotification(ctx context.Context, in *engine.SearchUserNotificationRequest) (*engine.ListUserNotification, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchUserNotification{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		CreatedAt: toModelFilterBetween(in.GetCreatedAt()),
		Read:      in.Read,
	}

	list, endList, err := api.ctrl.SearchUserNotification(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.UserNotification, 0, len(list))
	for _, v := range list {
		items = append(items, transformUserNotification(v))
	}

	return &engine.ListUserNotification{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *userNotification) CountUserNotification(ctx context.Context, in *engine.CountUserNotificationRequest) (*engine.CountUserNotificationResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchUserNotification{
		ListRequest: model.ListRequest{
			Q: in.GetQ(),
		},
		CreatedAt: toModelFilterBetween(in.GetCreatedAt()),
		Read:      in.Read,
	}

	count, err := api.ctrl.CountUserNotification(ctx, session, req)
	if err != nil {
		return nil, err
	}

	return &engine.CountUserNotificationResponse{
		Count: &count,
	}, nil
}

func (api *userNotification) MarkReadUserNotification(ctx context.Context, in *engine.MarkReadUserNotificationRequest) (*engine.MarkReadUserNotificationResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	count, err := api.ctrl.MarkReadUserNotification(ctx, session, in.GetId(), in.GetAll())
	if err != nil {
		return nil, err
	}

	return &engine.MarkReadUserNotificationResponse{
		Count: &count,
	}, nil
}

func toModelFilterBetween(src *engine.FilterBetween) *model.FilterBetween {
	if src == nil {
		return nil
	}

	return &model.FilterBetween{
		From: src.GetFrom(),
		To:   src.GetTo(),
	}
}

func transformUserNotification(src *model.UserNotification) *engine.UserNotification {
	return &engine.UserNotification{
		Id:        src.Id,
		CreatedAt: model.TimeToInt64(src.CreatedAt),
		Type:      src.Type,
		Message:   src.Message,
		ReadAt:    model.TimeToInt64(src.ReadAt),
	}
}
