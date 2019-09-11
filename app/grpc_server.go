package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	HEADER_TOKEN = strings.ToLower(model.HEADER_TOKEN)
)

type GrpcServer struct {
	srv *grpc.Server
}

func unaryInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	h, err := handler(ctx, req)

	if err != nil {
		wlog.Debug(fmt.Sprintf("method %s duration %s, error: %v", info.FullMethod, time.Since(start), err.Error()))

		switch err.(type) {
		case *model.AppError:
			e := err.(*model.AppError)
			return h, status.Error(httpCodeToGrpc(e.StatusCode), e.ToJson())
		default:
			return h, err
		}
	} else {
		wlog.Debug(fmt.Sprintf("method %s duration %s", info.FullMethod, time.Since(start)))
	}

	return h, err
}

func httpCodeToGrpc(c int) codes.Code {
	switch c {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusAccepted:
		return codes.ResourceExhausted
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	default:
		return codes.Internal

	}
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{
		srv: grpc.NewServer(
			grpc.UnaryInterceptor(unaryInterceptor),
		),
	}
}

func (s *GrpcServer) Server() *grpc.Server {
	return s.srv
}

func (a *App) StartGrpcServer() error {
	lis, err := net.Listen("tcp", "10.10.10.25:8081")
	if err != nil {
		panic(err.Error())
	}

	go func() {
		err := a.GrpcServer.srv.Serve(lis)
		if err != nil {
			panic(err.Error())
		}
	}()

	return nil
}

func (a *App) GetSessionFromCtx(ctx context.Context) (*model.Session, *model.AppError) {
	var session *model.Session
	var err *model.AppError
	var token []string

	if info, ok := metadata.FromIncomingContext(ctx); !ok {
		return nil, model.NewAppError("GetSessionFromCtx", "app.grpc.get_context", nil, "Not found", http.StatusInternalServerError)
	} else {
		token = info.Get(HEADER_TOKEN)
	}

	if len(token) < 1 {
		return nil, model.NewAppError("GetSessionFromCtx", "api.context.session_expired.app_error", nil, "token not found", http.StatusUnauthorized)
	}

	session, err = a.sessionManager.GetSession(token[0])
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		return nil, model.NewAppError("GetSessionFromCtx", "api.context.session_expired.app_error", nil, "token="+token[0], http.StatusUnauthorized)
	}

	return session, nil
}
