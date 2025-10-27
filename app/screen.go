package app

import (
	"context"
	"github.com/webitel/engine/model"
	"strconv"
	"strings"
)

const (
	requestScreenShare = "screen_share"
	requestScreenshot  = "screenshot"

	screenShareInvite = "invite"
	screenShareAccept = "accept"

	screenRecordStart = "ss_record_start"
	screenRecordStop  = "ss_record_stop"
	screenRecordACK   = "ss_ack"

	appDescTrack = "desc_track"
)

func (app *App) RequestScreenShare(ctx context.Context, domainId int64, fromUserId, toUserId int64, sockId string, sdp, id, ackId string) (string, model.AppError) {
	toSockId, err := app.Store.SocketSession().SockIdByApp(ctx, domainId, toUserId, appDescTrack)
	if err != nil {
		return "", err
	}

	if toSockId == "" {
		return "", model.NewNotFoundError("app.request_screen_share", "not found session")
	}

	err = app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    requestScreenShare,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		SockId:    &toSockId,
		Body: map[string]interface{}{
			"state":        screenShareInvite,
			"sock_id":      sockId,
			"parent_id":    id,
			"from_user_id": fromUserId,
			"auto":         true,
			"sdp":          sdp,
			"ack_id":       ackId,
		},
	})
	return sockId, err
}

func (app *App) AcceptScreenShare(domainId int64, toUserId int64, sockId string, sess, sdp string, fromSockId, ackId string) model.AppError {
	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    requestScreenShare,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		SockId:    &sockId,
		Body: map[string]interface{}{
			"state":        screenShareAccept,
			"sdp":          sdp,
			"session_id":   sess,
			"from_sock_id": fromSockId,
			"ack_id":       ackId,
		},
	})
}

func (app *App) Screenshot(ctx context.Context, domainId int64, toUserId int64, fromSockId string, ackId string) model.AppError {
	sockId, err := app.Store.SocketSession().SockIdByApp(ctx, domainId, toUserId, appDescTrack)
	if err != nil {
		return err
	}

	if sockId == "" {
		return model.NewNotFoundError("app.screenshot", "not found session")
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    requestScreenshot,
		SockId:    &sockId,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		Body: map[string]any{
			"from_sock_id": fromSockId,
			"ack_id":       ackId,
		},
	})
}

func (app *App) ScreenShareRecordStart(ctx context.Context, domainId int64, fromUserId, toUserId int64, rootSessionId, ackId string) model.AppError {
	sockId, err := app.Store.SocketSession().SockIdByApp(ctx, domainId, toUserId, appDescTrack)
	if err != nil {
		return err
	}

	if sockId == "" {
		return model.NewNotFoundError("app.start_record", "not found session")
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    screenRecordStart,
		SockId:    &sockId,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		Body: map[string]interface{}{
			"root_id":      rootSessionId,
			"from_user_id": fromUserId,
			"ack_id":       ackId,
		},
	})
}

func (app *App) ScreenShareRecordStop(ctx context.Context, domainId int64, toUserId int64, rootSessionId, ackId string) model.AppError {
	sockId, err := app.Store.SocketSession().SockIdByApp(ctx, domainId, toUserId, appDescTrack)
	if err != nil {
		return err
	}

	if sockId == "" {
		return model.NewNotFoundError("app.stop_record", "not found session")
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    screenRecordStop,
		SockId:    &sockId,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		Body: map[string]interface{}{
			"root_id": rootSessionId,
			"ack_id":  ackId,
		},
	})
}

func (app *App) ACK(ctx context.Context, domainId int64, ackId, errText string) model.AppError {
	var sockId string
	var toUserId int

	arr := strings.SplitN(ackId, "/", 3)
	if len(arr) != 3 {
		return model.NewBadRequestError("app.ack", "bad ack")
	}
	b := map[string]any{
		"ack_id": ackId,
	}

	if errText != "" {
		b["error"] = errText
	}

	sockId = arr[2]
	toUserId, _ = strconv.Atoi(arr[1])

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    screenRecordACK,
		SockId:    &sockId,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{int64(toUserId)},
		Body:      b,
	})
}
