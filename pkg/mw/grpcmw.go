package mw

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var traceID string
		var userAgent string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-trace-id"); len(ids) > 0 {
				traceID = ids[0]
			}
			if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
				userAgent = userAgents[0]
			}
		}

		if traceID == "" {
			traceID = uuid.New().String()
		}

		clientIP := "undefined"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceID)

		startTime := time.Now()

		logger.Info("got request in grpc",
			slog.String("trace_id", traceID),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("method", info.FullMethod),
			slog.Int64("timestamp", startTime.Unix()))

		response, err := handler(ctx, req)

		duration := time.Since(startTime)
		statusCode, _ := status.FromError(err)

		logger.Info("request compete in grpc",
			slog.String("trace_id", traceID),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("method", info.FullMethod),
			slog.String("status_code", statusCode.Code().String()),
			slog.Duration("process_time", duration),
			slog.Int64("timestamp", time.Now().Unix()))

		return response, err
	}
}

func StreamServerInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(server any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		var traceID string
		var userAgent string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-trace-id"); len(ids) > 0 {
				traceID = ids[0]
			}
			if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
				userAgent = userAgents[0]
			}
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		clientIP := "undefined"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		startTime := time.Now()

		logger.Info("got stream request in grpc",
			slog.String("trace_id", traceID),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("method", info.FullMethod),
			slog.Bool("client_stream", info.IsClientStream),
			slog.Bool("server_stream", info.IsServerStream),
			slog.Int64("timestamp", startTime.Unix()))

		err := handler(server, ss)

		duration := time.Since(startTime)
		statusCode, _ := status.FromError(err)

		logger.Info("request compete in grpc",
			slog.String("trace_id", traceID),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("method", info.FullMethod),
			slog.String("status_code", statusCode.Code().String()),
			slog.Duration("process_time", duration),
			slog.Int64("timestamp", time.Now().Unix()))

		return err
	}
}
