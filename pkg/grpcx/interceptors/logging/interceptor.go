package logging

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"webooktrial/pkg/grpcx/interceptors"
	"webooktrial/pkg/logger"
)

type InterceptorBuilder struct {
	// 如果要非常通用
	l logger.LoggerV1
	//fn func(msg string, fields...logger.Field)
	interceptors.Builder
	reqBody  bool
	respBody bool
}

func NewInterceptorBuilder(l logger.LoggerV1) *InterceptorBuilder {
	return &InterceptorBuilder{l: l}
}

func (b *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//start := time.Now()
		//event := "normal"
		//defer func() {
		//	// 照着抄
		//}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (b *InterceptorBuilder) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// 默认过滤掉该探活日志
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		var start = time.Now()
		var fields = make([]logger.Field, 0, 20)
		var event = "normal"

		defer func() {
			cost := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}
			st, _ := status.FromError(err)
			fields = append(fields,
				logger.String("type", "unary"),
				logger.String("code", st.Code().String()),
				logger.String("code_msg", st.Message()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_ip", b.PeerIP(ctx)),
			)
			b.l.Info("RPC调用", fields...)
		}()

		return handler(ctx, req)
	}
}
