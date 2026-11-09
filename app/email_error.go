package app

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/webitel/engine/model"
)

// emailStableId matches the stable ids call_center sends as the status message.
var emailStableId = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$`)

var emailHttpCodes = map[codes.Code]int{
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.FailedPrecondition: http.StatusBadRequest,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.NotFound:           http.StatusNotFound,
	codes.Aborted:            http.StatusConflict,
	codes.AlreadyExists:      http.StatusConflict,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.Canceled:           499,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
}

// emailGrpcError keeps the stable id of call_center and the code of its status; free text gets
// the fallback id of the code, and the end of the caller's context keeps its own code.
func emailGrpcError(ctx context.Context, err error) model.AppError {
	code := codes.Unknown
	message := ""
	if st, ok := status.FromError(err); ok {
		code, message = st.Code(), st.Message()
	}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		code = codes.DeadlineExceeded
	case errors.Is(ctx.Err(), context.Canceled):
		code = codes.Canceled
	}

	httpCode, known := emailHttpCodes[code]
	if !known {
		code, httpCode = codes.Internal, http.StatusInternalServerError
	}
	id := message
	if !emailStableId.MatchString(id) {
		id = "engine.email_attempt." + snakeCase(code.String())
	}

	return model.NewCustomCodeError(id, message, httpCode)
}

// snakeCase turns a gRPC code name such as InvalidArgument into invalid_argument.
func snakeCase(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}

	return b.String()
}
