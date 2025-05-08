package app

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
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

const (
	traceparentHeader = "micro-trace-id"
	tracestateHeader  = "micro-span-id"
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

func ExtractContextFromMessageAttributes(ctx context.Context) context.Context {
	attributes := make(map[string]string)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	if val, ok := md[traceparentHeader]; ok {
		attributes["traceparent"] = val[0]
	}
	if val, ok := md[tracestateHeader]; ok {
		attributes["tracestate"] = val[0]
	}

	if len(attributes) == 0 {
		return ctx
	}

	return propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier(attributes))
}

// TODO sync map ?
type GrpcHeaderCarrier map[string][]string

// Get returns the value associated with the passed key.
func (hc GrpcHeaderCarrier) Get(key string) string {
	if v, ok := hc[key]; ok && len(v) != 0 {
		return v[0]
	}
	return ""
}

// Set stores the key-value pair.
func (hc GrpcHeaderCarrier) Set(key string, value string) {
	hc[key] = []string{value}
}

// Keys lists the keys stored in this carrier.
func (hc GrpcHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

func GetUnaryInterceptor(app *App) grpc.UnaryServerInterceptor {
	tc := app.Tracer()
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		var sess *auth_manager.Session
		var md metadata.MD

		start := time.Now()

		md, sess, err = app.getSessionFromCtx(ctx)
		if err != nil {
			app.Log.Error(err.Error(), wlog.Err(err))
			sess = &auth_manager.Session{}
		}

		var reqCtx context.Context

		if md == nil {
			md = metadata.MD{}
		}

		propagators := otel.GetTextMapPropagator()
		ctx = propagators.Extract(
			ctx, GrpcHeaderCarrier(md),
		)

		spanCtx, span := tc.Start(ctx, info.FullMethod)
		defer func() {
			span.End()
		}()

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

	if !ok || sess.UserId == 0 {
		return nil, model.NewUnauthorizedError("session.valid", "Unauthenticated")
	}
	return sess, nil
}

func (a *App) getSessionFromCtx(ctx context.Context) (metadata.MD, *auth_manager.Session, model.AppError) {
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
		return info, nil, model.NewUnauthorizedError("app.grpc.get_context", "Not found")
	} else {
		token = info.Get(HEADER_TOKEN)
	}

	if len(token) < 1 {
		return info, nil, model.NewUnauthorizedError("api.context.session_expired.app_error", "token not found")
	}

	session, err = a.GetSession(token[0])
	if err != nil {
		return info, nil, err
	}

	if session.IsExpired() {
		return info, nil, model.NewUnauthorizedError("api.context.session_expired.app_error", "token="+token[0])
	}

	session.SetIp(getClientIp(info))

	return info, session, nil
}

func getClientIp(info metadata.MD) string {
	ip := strings.Join(info.Get("x-real-ip"), ",")
	if ip == "" {
		ip = strings.Join(info.Get("x-forwarded-for"), ",")
	}

	return ip
}
