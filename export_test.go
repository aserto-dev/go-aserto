package aserto

import (
	"google.golang.org/grpc"
)

func InternalNewConnection(dialContext connectionFactory, options *ConnectionOptions) (*grpc.ClientConn, error) {
	return newConnection(dialContext, options)
}

func InternalUnary(tenantID, sessionID string) grpc.UnaryClientInterceptor {
	return unary(tenantID, sessionID)
}

func InternalStream(tenantID, sessionID string) grpc.StreamClientInterceptor {
	return stream(tenantID, sessionID)
}
