package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// кольцевой буфер журнала процесса: пишут интерцепторы, читает rpc GetLogs
var requestLogs = NewLogRing(500)

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, "internal error")
				log.Println(err, r)
				requestLogs.Add(LogEntry{
					AtMs: nowMs(), Level: "ERROR", Path: info.FullMethod,
					Code: "Internal", Message: "panic recovered",
				})
			}
		}()
		return handler(ctx, req)
	}
}

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			code := status.Code(err)
			log.Printf("method=%s duration=%s err=%v", info.FullMethod, duration, err)
			requestLogs.Add(LogEntry{
				AtMs:  nowMs(),
				Level: levelForGRPCCode(code.String()),
				Path:  info.FullMethod,
				Code:  code.String(),
				MS:    duration.Milliseconds(),
			})
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}
