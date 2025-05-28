package app

import "github.com/webitel/engine/model"

const (
	requestScreenShare = "screen_share"

	screenShareInvite = "invite"
	screenShareAccept = "accept"
)

func (app *App) RequestScreenShare(domainId int64, fromUserId, toUserId int64, sockId string, sdp, id string) model.AppError {
	err := app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    requestScreenShare,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		Body: map[string]interface{}{
			"state":        screenShareInvite,
			"sock_id":      sockId,
			"parent_id":    id,
			"from_user_id": fromUserId,
			"auto":         true,
			"timeout":      10000,
			"sdp":          sdp,
		},
	})
	return err
}

func (app *App) AcceptScreenShare(domainId int64, toUserId int64, sockId string, sess, sdp string) model.AppError {
	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    requestScreenShare,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{toUserId},
		SockId:    &sockId,
		Body: map[string]interface{}{
			"state":      screenShareAccept,
			"sdp":        sdp,
			"session_id": sess,
		},
	})
}
