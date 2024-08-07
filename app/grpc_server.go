package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/localization"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/utils"
	"github.com/webitel/wlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	HEADER_TOKEN = strings.ToLower(model.HEADER_TOKEN)

	RequestContextName = "grpc_ctx"
)

type GrpcServer struct {
	srv *grpc.Server
	lis net.Listener
}

func (grpc *GrpcServer) GetPublicInterface() (string, int) {
	h, p, _ := net.SplitHostPort(grpc.lis.Addr().String())
	if h == "::" {
		h = utils.GetPublicAddr()
	}
	port, _ := strconv.Atoi(p)
	return h, port
}

func unaryInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	var reqCtx context.Context
	var ip string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		reqCtx = context.WithValue(ctx, RequestContextName, md)
		ip = getClientIp(md)
	} else {
		ip = "<not found>"
		reqCtx = context.WithValue(ctx, RequestContextName, nil)
	}

	log := wlog.GlobalLogger().With(wlog.Namespace("context"),
		//wlog.Int64("domain_id", -1),
		//wlog.Int64("user_id", -1),
		wlog.String("ip_address", ip),
		wlog.String("method", info.FullMethod),
	)

	h, err := handler(reqCtx, req)

	if err != nil {
		log.Error(err.Error(), wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/float64(1000)))

		switch err.(type) {
		case model.AppError:
			e := err.(model.AppError)
			e.Translate(localization.TfuncWithFallback(model.DEFAULT_LOCALE))
			return h, status.Error(httpCodeToGrpc(e.GetStatusCode()), e.ToJson())
		default:
			return h, err
		}
	} else {
		log.Debug("ok", wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/float64(1000)))
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

func NewGrpcServer(settings model.ServerSettings) *GrpcServer {
	address := fmt.Sprintf("%s:%d", settings.Address, settings.Port)
	lis, err := net.Listen(settings.Network, address)
	if err != nil {
		panic(err.Error())
	}

	return &GrpcServer{
		lis: lis,
		srv: grpc.NewServer(
			grpc.UnaryInterceptor(unaryInterceptor),
			grpc.MaxRecvMsgSize(int(settings.MaxMessageSize)),
			grpc.MaxSendMsgSize(int(settings.MaxMessageSize)),
		),
	}
}

func (s *GrpcServer) Server() *grpc.Server {
	return s.srv
}

func (a *App) StartGrpcServer() error {
	go func() {
		defer wlog.Debug(fmt.Sprintf("[grpc] close server listening"))
		wlog.Debug(fmt.Sprintf("[grpc] server listening %s", a.GrpcServer.lis.Addr().String()))
		err := a.GrpcServer.srv.Serve(a.GrpcServer.lis)
		if err != nil {
			//FIXME
			panic(err.Error())
		}
	}()

	return nil
}

func (a *App) GetSessionFromCtx(ctx context.Context) (*auth_manager.Session, model.AppError) {
	var session *auth_manager.Session
	var err model.AppError
	var token []string
	var info metadata.MD
	var ok bool

	v := ctx.Value(RequestContextName)
	info, ok = v.(metadata.MD)

	// todo
	if !ok {
		info, ok = metadata.FromIncomingContext(ctx)
	}

	if !ok {
		return nil, model.NewUnauthorizedError("app.grpc.get_context", "Not found")
	} else {
		token = info.Get(HEADER_TOKEN)
	}

	if len(token) < 1 {
		return nil, model.NewUnauthorizedError("api.context.session_expired.app_error", "token not found")
	}

	session, err = a.GetSession(token[0])
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		return nil, model.NewUnauthorizedError("api.context.session_expired.app_error", "token="+token[0])
	}

	session.SetIp(getClientIp(info))

	return session, nil
}

func getClientIp(info metadata.MD) string {
	ip := strings.Join(info.Get("x-real-ip"), ",")
	if ip == "" {
		ip = strings.Join(info.Get("x-forwarded-for"), ",")
	}

	return ip
}
