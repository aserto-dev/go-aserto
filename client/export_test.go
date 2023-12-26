package client

import (
	"context"

	"google.golang.org/grpc"
)

func InternalNewConnection(ctx context.Context,
	dialContext dialer,
	options *ConnectionOptions,
) (*grpc.ClientConn, error) {
	return newConnection(ctx, dialContext, options)
}

func InternalUnary(tenantID, sessionID string) grpc.UnaryClientInterceptor {
	return unary(tenantID, sessionID)
}

func InternalStream(tenantID, sessionID string) grpc.StreamClientInterceptor {
	return stream(tenantID, sessionID)
}
