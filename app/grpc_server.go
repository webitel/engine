package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/localization"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
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

	RequestContextName    = "grpc_ctx"
	RequestContextSession = "session"
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

func GetUnaryInterceptor(app *App) grpc.UnaryServerInterceptor {
	tp := otel.GetTracerProvider()
	tc := tp.Tracer("engine")

	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		var sess *auth_manager.Session

		start := time.Now()

		sess, err = app.getSessionFromCtx(ctx)
		if err != nil {
			return nil, err
		}

		var reqCtx context.Context

		spanCtx, span := tc.Start(ctx, info.FullMethod)
		defer span.End()

		span.SetAttributes(
			attribute.Int64("domain_id", sess.DomainId),
			attribute.Int64("user_id", sess.UserId),
			attribute.String("ip_address", sess.GetUserIp()),
			attribute.String("method", info.FullMethod),
		)

		reqCtx = context.WithValue(spanCtx, RequestContextSession, sess)
		log := app.Log.With(wlog.Namespace("context"),
			wlog.Int64("domain_id", sess.DomainId),
			wlog.Int64("user_id", sess.UserId),
			wlog.String("ip_address", sess.GetUserIp()),
			wlog.String("method", info.FullMethod),
		)

		var h any
		h, err = handler(reqCtx, req)

		if err != nil {
			log.Error(err.Error(), wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/float64(1000)))
			span.SetStatus(otelCodes.Error, err.Error())

			switch err.(type) {
			case model.AppError:
				e := err.(model.AppError)
				e.Translate(localization.TfuncWithFallback(model.DEFAULT_LOCALE))
				return h, status.Error(httpCodeToGrpc(e.GetStatusCode()), e.ToJson())
			default:
				return h, err
			}
		} else {
			span.SetStatus(otelCodes.Ok, "success")
			log.Debug("200", wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/float64(1000)))
		}

		return h, err
	}
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

func NewGrpcServer(app *App, settings model.ServerSettings) *GrpcServer {
	address := fmt.Sprintf("%s:%d", settings.Address, settings.Port)
	lis, err := net.Listen(settings.Network, address)
	if err != nil {
		panic(err.Error())
	}

	return &GrpcServer{
		lis: lis,
		srv: grpc.NewServer(
			//grpc.StatsHandler(otelgrpc.NewServerHandler()),
			//grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.UnaryInterceptor(GetUnaryInterceptor(app)),
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
	v := ctx.Value(RequestContextSession)
	sess, ok := v.(*auth_manager.Session)

	// todo
	if !ok {

	}
	return sess, nil
}

func (a *App) getSessionFromCtx(ctx context.Context) (*auth_manager.Session, model.AppError) {
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
