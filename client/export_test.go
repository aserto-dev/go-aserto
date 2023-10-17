package client

import (
	"context"

	"google.golang.org/grpc"
)

func InternalNewConnection(ctx context.Context,
	dialContext dialer,
	options *ConnectionOptions,
) (*Connection, error) {
	return newConnection(ctx, dialContext, options)
}

func (c *Connection) InternalUnary(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return c.unary(ctx, method, req, reply, cc, invoker, opts...)
}

func (c *Connection) InternalStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return c.stream(ctx, desc, cc, method, streamer, opts...)
}
