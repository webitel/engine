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

type config struct {
	dialOptions []grpc.DialOption
	lbPolicy    string
}

type Option func(*config)

func WithGrpcOptions(opts ...grpc.DialOption) Option {
	return func(c *config) {
		c.dialOptions = append(c.dialOptions, opts...)
	}
}

func WithLBPolicy(policy string) Option {
	return func(c *config) {
		c.lbPolicy = policy
	}
}

func NewClient[T any](consulTarget, service string, api func(grpc.ClientConnInterface) T, opts ...Option) (*Client[T], error) {

	cfg := &config{
		lbPolicy:    "wbt_round_robin",
		dialOptions: []grpc.DialOption{
			//grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if len(cfg.dialOptions) == 0 {
		cfg.dialOptions = append(cfg.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	dsn := fmt.Sprintf("wbt://%s/%s?wait=15s", consulTarget, service)

	dialOpts := append(cfg.dialOptions,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, cfg.lbPolicy)),
	)

	actual, _ := conns.LoadOrStore(dsn, func() interface{} {
		conn, err := grpc.NewClient(dsn, dialOpts...)
		if err != nil {
			return err // Тут треба бути обережним, краще зберігати структуру-обгортку
		}
		return conn
	}())

	conn, ok := actual.(*grpc.ClientConn)
	if !ok {
		return nil, fmt.Errorf("failed to create or retrieve connection")
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
	header := metadata.New(map[string]string{AuthHeaderName: token})
	return metadata.NewOutgoingContext(ctx, header)
}

func (c *Client[T]) Close() error {
	return c.conn.Close()
}

func (c *Client[T]) Start() error {
	return nil
}
