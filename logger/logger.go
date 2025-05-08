package logger

import (
	"context"
	"encoding/json"
	"fmt"
	proto "github.com/webitel/engine/gen/logger"
	"github.com/webitel/engine/pkg/wbt"
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
	service *wbt.Client[proto.ConfigServiceClient]
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

func New(consulTarget string, channel Publisher) (*Audit, error) {
	service, err := wbt.NewClient(consulTarget, loggerServiceName, proto.NewConfigServiceClient)
	if err != nil {
		return nil, err
	}

	return &Audit{
		service: service,
		cache:   utils.NewLruWithParams(sizeCache, "logger-cache", expires, ""),
		channel: channel,
	}, nil
}

func (api *Audit) checkIsActive(ctx context.Context, domainId int64, object string) (bool, error) {
	key := fmt.Sprintf("%d-%s", domainId, object)

	res, err, _ := group.Do(key, func() (interface{}, error) {
		if res, ok := api.cache.Get(key); ok {
			return res.(bool), nil
		}

		res, err := api.service.Api.CheckConfigStatus(ctx, &proto.CheckConfigStatusRequest{
			ObjectName: object,
			DomainId:   domainId,
		})

		if err != nil {
			return nil, err
		}

		api.cache.AddWithDefaultExpires(key, res.IsEnabled)
		return res.IsEnabled, nil
	})

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

func (api *Audit) Audit(action Action, ctx context.Context, session Session, object string, recordId int64, data interface{}) error {
	var body []byte
	ok, err := api.checkIsActive(ctx, session.GetDomainId(), object)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	body, err = json.Marshal(data)
	if err != nil {
		return err
	}

	msg := Message{
		Records: []Record{
			{
				Id:       recordId,
				NewState: body,
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

	err = api.channel.Send(ctx, exchange, fmt.Sprintf(rkFormat, msg.DomainId, object), msg.ToJson())
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (api *Audit) Create(ctx context.Context, session Session, object string, recordId int64, data interface{}) error {
	return api.Audit(ActionCreate, ctx, session, object, recordId, data)
}

func (api *Audit) Update(ctx context.Context, session Session, object string, recordId int64, data interface{}) error {
	return api.Audit(ActionUpdate, ctx, session, object, recordId, data)
}

func (api *Audit) Delete(ctx context.Context, session Session, object string, recordId int64, data interface{}) error {
	return api.Audit(ActionDelete, ctx, session, object, recordId, data)
}
