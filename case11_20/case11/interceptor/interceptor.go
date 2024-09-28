package interceptor

import (
	"context"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor(tb *TokenBucket) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !tb.Consume(1) {
			ctx = context.WithValue(ctx, "RateLimited", true)
		}
		// 继续处理请求
		return handler(ctx, req)
	}
}
