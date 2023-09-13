package logger

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/webitel/engine/utils"
	proto "github.com/webitel/protos/logger"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"time"
)

const (
	sizeCache = 10 * 1000
	expires   = 10 * 1000 // 10s

	exchange = "logger"
	rkFormat = "logger.%d.%s"
)

var (
	group singleflight.Group
)

type Audit struct {
	service proto.ConfigServiceClient
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
	conn, err := grpc.Dial(fmt.Sprintf("consul://%s/logger?wait=14s", consulTarget),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	service := proto.NewConfigServiceClient(conn)

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

		res, err := api.service.CheckConfigStatus(ctx, &proto.CheckConfigStatusRequest{
			ObjectName: object,
			DomainId:   domainId,
		})

		if err != nil {
			return nil, err
		}

		api.cache.Add(key, res.IsEnabled)
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
