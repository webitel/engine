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
)

const (
	sizeCache = 10 * 1000
	expires   = 10 * 1000 // 10s
)

var (
	group singleflight.Group
)

type Api struct {
	service proto.ConfigServiceClient
	cache   *utils.Cache
	channel Publisher
}

type AuditRec struct {
	DomainId int64
	Object   string
}

type Publisher interface {
	Send(ctx context.Context, domainId int64, object string, body []byte) error
}

func New(consulTarget string, channel Publisher) (*Api, error) {
	conn, err := grpc.Dial(fmt.Sprintf("consul://%s/logger?wait=14s", consulTarget),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	service := proto.NewConfigServiceClient(conn)

	return &Api{
		service: service,
		cache:   utils.NewLruWithParams(sizeCache, "logger-cache", expires, ""),
		channel: channel,
	}, nil
}

func (api *Api) checkIsActive(ctx context.Context, domainId int64, object string) (bool, error) {
	key := fmt.Sprintf("%d-%s", domainId, object)

	res, err, _ := group.Do(key, func() (interface{}, error) {
		if res, ok := api.cache.Get(key); ok {
			return res.(bool), nil
		}

		res, err := api.service.ReadConfigByObjectId(ctx, &proto.ReadConfigByObjectIdRequest{
			ObjectId: int32(259439),   // TODO
			DomainId: int32(domainId), // TODO
		})

		if err != nil {
			return nil, err
		}

		api.cache.Add(key, res.GetEnabled())
		return res.GetEnabled(), nil
	})

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

func (api *Api) Audit(ctx context.Context, domainId int64, object string, data interface{}) error {
	var body []byte
	ok, err := api.checkIsActive(ctx, domainId, object)
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

	return api.channel.Send(ctx, domainId, object, body)
}
