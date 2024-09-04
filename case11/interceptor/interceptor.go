package interceptor

import (
	"context"
	ratelimiter "interview-cases/case11/ratelimit"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(tb *ratelimiter.TokenBucket) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !tb.Consume(1) {
			ctx = context.WithValue(ctx, "RateLimited", true)
		}

		// 继续处理请求
		return handler(ctx, req)
	}
}
