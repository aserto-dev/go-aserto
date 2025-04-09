package mock

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

var errNotImplemented = errors.New("not implemented")

// Mock grpc.ServerStream.
type ServerStream struct{}

func (s *ServerStream) SetHeader(metadata.MD) error {
	return errNotImplemented
}

func (s *ServerStream) SendHeader(metadata.MD) error {
	return errNotImplemented
}

func (s *ServerStream) SetTrailer(metadata.MD) {
}

func (s *ServerStream) Context() context.Context {
	return context.TODO()
}

func (s *ServerStream) SendMsg(_ any) error {
	return errNotImplemented
}

func (s *ServerStream) RecvMsg(_ any) error {
	return errNotImplemented
}
