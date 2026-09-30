package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const ForwardedUserAgentHeader string = "x-forwarded-user-agent"

func UserAgentFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	for _, ua := range md.Get(ForwardedUserAgentHeader) {
		if ua != "" {
			return ua, true
		}
	}

	return "", false
}
