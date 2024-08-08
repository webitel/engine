package utils

import (
	"time"

	otelgrpc "github.com/webitel/webitel-go-kit/tracing/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClientConn(endpoint string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// TODO: remove deprecated func
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithMessageEvents(otelgrpc.SentEvents, otelgrpc.ReceivedEvents))),
		grpc.WithBlock(), grpc.WithTimeout(timeout))

	return grpc.Dial(endpoint, opts...)
}
