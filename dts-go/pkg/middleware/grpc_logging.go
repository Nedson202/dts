package middleware

import (
	"context"
	"time"

	"github.com/nedson202/dts-go/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		st, _ := status.FromError(err)
		logger.Info().
			Str("method", info.FullMethod).
			Dur("duration", duration).
			Interface("request", req).
			Int("code", int(st.Code())).
			Msg("gRPC request")

		return resp, err
	}
}
