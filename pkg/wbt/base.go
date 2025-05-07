package wbt

import (
	"context"
	"fmt"
	"github.com/webitel/engine/pkg/wbt/consul"
	_ "github.com/webitel/engine/pkg/wbt/consul"
	"google.golang.org/grpc"
)

//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.fs.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.webitel.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.engine.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.cc.yaml
//go:generate go mod tidy

type Client[T any] struct {
	conn *grpc.ClientConn
	Api  T
}

func NewClient[T any](consulTarget string, service string, api func(conn grpc.ClientConnInterface) T) (*Client[T], error) {
	conn, err := grpc.Dial(fmt.Sprintf("wbt://%s/%s?wait=15s", consulTarget, service),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "wbt_round_robin"}`),
		grpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	return &Client[T]{
		conn: conn,
		Api:  api(conn),
	}, nil
}

func (c *Client[T]) StaticHost(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, consul.StaticHostKey, consul.StaticHost{Name: name})
}

func (c *Client[T]) Close() error {
	return c.conn.Close()
}

func (c *Client[T]) Start() error {
	return nil
}
