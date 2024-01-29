package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger Logger
}

func NewLoggingInterceptor(logg Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logg,
	}
}

func (l *LoggingInterceptor) UnaryServerLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	initTime := time.Now()
	resp, err := handler(ctx, req)
	latency := time.Since(initTime)

	serverLog(ctx, l.logger, info, err, latency)

	return resp, err
}

func serverLog(
	ctx context.Context,
	logger Logger,
	info *grpc.UnaryServerInfo,
	err error,
	latency time.Duration,
) {
	p, ok := peer.FromContext(ctx)

	var clientIP string
	if ok {
		clientIP = p.Addr.String()
	} else {
		clientIP = ""
	}

	meta, ok := metadata.FromIncomingContext(ctx)

	var userAgent string
	if ok {
		userAgent = meta["user-agent"][0]
	} else {
		userAgent = ""
	}

	statusCode := 0
	if err != nil {
		statusCode = int(status.Code(err))
	}

	logger.Info(fmt.Sprintf(
		"%s %s %d %s \"%s\"",
		clientIP,
		info.FullMethod,
		statusCode,
		latency,
		userAgent,
	))
}
