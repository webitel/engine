package wbt

import (
	"context"
	"fmt"
	"github.com/webitel/engine/pkg/wbt/consul"
	_ "github.com/webitel/engine/pkg/wbt/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"sync"
)

//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.fs.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.webitel.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.engine.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.cc.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.chat.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.flow.yaml
//go:generate go mod tidy

type Client[T any] struct {
	conn *grpc.ClientConn
	Api  T
}

var conns sync.Map

func NewClient[T any](consulTarget string, service string, api func(conn grpc.ClientConnInterface) T) (*Client[T], error) {
	var conn *grpc.ClientConn
	var err error

	dsn := fmt.Sprintf("wbt://%s/%s?wait=15s", consulTarget, service)

	if c, ok := conns.Load(dsn); ok {
		conn = c.(*grpc.ClientConn)
	} else {
		conn, err = grpc.NewClient(dsn,
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "wbt_round_robin"}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			return nil, err
		}
		conns.Store(dsn, conn)
	}

	return &Client[T]{
		conn: conn,
		Api:  api(conn),
	}, nil
}

func (c *Client[T]) StaticHost(ctx context.Context, name string) context.Context {
	return StaticHost(ctx, name)
}

func (c *Client[T]) WithToken(ctx context.Context, token string) context.Context {
	return WithToken(ctx, token)
}

func StaticHost(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, consul.StaticHostKey, consul.StaticHost{Name: name})
}

func WithToken(ctx context.Context, token string) context.Context {
	header := metadata.New(map[string]string{"x-webitel-access": token})
	return metadata.NewOutgoingContext(ctx, header)
}

func (c *Client[T]) Close() error {
	return c.conn.Close()
}

func (c *Client[T]) Start() error {
	return nil
}
