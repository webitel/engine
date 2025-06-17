package logger

import (
	"context"
	"fmt"
	"github.com/webitel/engine/utils"
	"golang.org/x/sync/singleflight"
	"time"
)

const (
	sizeCache = 10 * 1000
	expires   = 10

	exchange          = "logger"
	rkFormat          = "logger.%d.%s"
	loggerServiceName = "logger"
)

var (
	group singleflight.Group
)

type Audit struct {
	cache   *utils.Cache
	channel Publisher
}

type AuditRec struct {
	DomainId int64
	Object   string
}

type Session interface {
	GetUserId() int64
	GetUserIp() string
	GetDomainId() int64
}

type Publisher interface {
	Send(ctx context.Context, exchange string, rk string, body []byte) error
}

func New(channel Publisher) (*Audit, error) {
	return &Audit{
		channel: channel,
	}, nil
}

func (api *Audit) Audit(action Action, ctx context.Context, session Session, object string, recordId string, data interface{}) error {
	msg := Message{
		Records: []Record{
			{
				Id:       recordId,
				NewState: data,
			},
		},
		RequiredFields: RequiredFields{
			UserId:     session.GetUserId(),
			UserIp:     session.GetUserIp(),
			DomainId:   session.GetDomainId(),
			Action:     action,
			Date:       time.Now().Unix(),
			ObjectName: object,
		},
	}

	err := api.channel.Send(ctx, exchange, fmt.Sprintf(rkFormat, msg.DomainId, object), msg.ToJson())
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (api *Audit) Create(ctx context.Context, session Session, object string, recordId string, data interface{}) error {
	return api.Audit(ActionCreate, ctx, session, object, recordId, data)
}

func (api *Audit) Update(ctx context.Context, session Session, object string, recordId string, data interface{}) error {
	return api.Audit(ActionUpdate, ctx, session, object, recordId, data)
}

func (api *Audit) Delete(ctx context.Context, session Session, object string, recordId string, data interface{}) error {
	return api.Audit(ActionDelete, ctx, session, object, recordId, data)
}
