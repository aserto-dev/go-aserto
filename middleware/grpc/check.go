package grpc

import (
	"context"
	"strings"

	"github.com/aserto-dev/go-aserto/middleware/internal"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	ds3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	DefaultSubjType = "user"
	DefaultObjType  = "tenant"
)

type Mapper func(ctx context.Context, req interface{}) (id string)

type CheckMiddleware struct {
	dsReader             ds3.ReaderClient
	subjType             string
	objType              string
	defaultObjType       string
	defaultObjID         string
	subjMapper           Mapper
	objMapper            Mapper
	permissionFromMethod bool
	ignoredMethods       []string
	ignoreCtx            map[interface{}][]string
}

func (c *CheckMiddleware) WithSubjectType(value string) *CheckMiddleware {
	c.subjType = value
	return c
}

func (c *CheckMiddleware) WithObjectType(value string) *CheckMiddleware {
	c.objType = value
	return c
}

func (c *CheckMiddleware) WithDefaultObjectType(value string) *CheckMiddleware {
	c.defaultObjType = value
	return c
}

func (c *CheckMiddleware) WithDefaultObjectID(value string) *CheckMiddleware {
	c.defaultObjID = value
	return c
}

func (c *CheckMiddleware) WithPermissionFromMethod() *CheckMiddleware {
	c.permissionFromMethod = true
	return c
}

func (c *CheckMiddleware) WithSubjectFromContextValue(ctxKey interface{}) *CheckMiddleware {
	c.subjMapper = func(ctx context.Context, _ interface{}) string {
		return internal.ValueOrEmpty(ctx, ctxKey)
	}

	return c
}

func (c *CheckMiddleware) WithObjectFromContextValue(ctxKey interface{}) *CheckMiddleware {
	c.objMapper = func(ctx context.Context, _ interface{}) string {
		return internal.ValueOrEmpty(ctx, ctxKey)
	}

	return c
}

func (c *CheckMiddleware) WithSubjectMapper(subjectMapper Mapper) *CheckMiddleware {
	c.subjMapper = subjectMapper
	return c
}

func (c *CheckMiddleware) WithObjectMapper(objectMapper Mapper) *CheckMiddleware {
	c.objMapper = objectMapper
	return c
}

func (c *CheckMiddleware) WithIgnoredMethods(methods []string) *CheckMiddleware {
	c.ignoredMethods = methods
	return c
}

func (c *CheckMiddleware) WithAutoAuthorizedContextValues(ctxKey interface{}, values []string) *CheckMiddleware {
	c.ignoreCtx[ctxKey] = values
	return c
}

func NewCheckMiddleware(reader ds3.ReaderClient) *CheckMiddleware {
	return &CheckMiddleware{
		dsReader:             reader,
		ignoredMethods:       []string{},
		permissionFromMethod: true,
		defaultObjType:       DefaultObjType,
		ignoreCtx:            map[interface{}][]string{},
	}
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := c.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		if err := c.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (c *CheckMiddleware) authorize(ctx context.Context, req interface{}) error {
	for ctxKey, values := range c.ignoreCtx {
		for _, value := range values {
			if internal.ValueOrEmpty(ctx, ctxKey) == value {
				return nil
			}
		}
	}

	objectID := c.objMapper(ctx, req)
	objectType := c.objectType()

	if objectID == "" {
		objectID = c.defaultObjID
		objectType = c.defaultObjType
	}

	if objectID == "" {
		return errors.New("object ID is empty")
	}

	subjectID := c.subjMapper(ctx, req)
	permission := ""

	if c.permissionFromMethod {
		permission = methodResource(ctx)
		for _, path := range c.ignoredMethods {
			if strings.EqualFold(path, permission) {
				return nil
			}
		}
	}

	allowed, err := c.dsReader.CheckPermission(ctx, &ds3.CheckPermissionRequest{
		SubjectType: c.subjectType(),
		SubjectId:   subjectID,
		ObjectType:  objectType,
		ObjectId:    objectID,
		Permission:  permission})
	if err != nil {
		return errors.Wrap(err, "failed to check permission for identity")
	}

	if !allowed.Check {
		return aerr.ErrAuthorizationFailed
	}

	return nil
}
func (c *CheckMiddleware) objectType() string {
	if c.objType == "" {
		return DefaultObjType
	}

	return c.objType
}

func (c *CheckMiddleware) subjectType() string {
	if c.subjType == "" {
		return DefaultSubjType
	}

	return c.subjType
}
