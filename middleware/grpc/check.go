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

type TypeIDMapper func(ctx context.Context, req interface{}) (objectType, objectID string)
type PermissionMapper func(ctx context.Context, req interface{}) (permission string)

type CheckMiddleware struct {
	dsReader         ds3.ReaderClient
	subjType         string
	objType          string
	subjMapper       TypeIDMapper
	objMapper        TypeIDMapper
	permissionMapper PermissionMapper
	ignoredMethods   []string
	ignoreCtx        map[interface{}][]string
}

func (c *CheckMiddleware) WithSubjectType(value string) *CheckMiddleware {
	c.subjType = value
	return c
}

func (c *CheckMiddleware) WithObjectType(value string) *CheckMiddleware {
	c.objType = value
	return c
}

func (c *CheckMiddleware) WithSubjectFromContextValue(ctxKey interface{}) *CheckMiddleware {
	c.subjMapper = func(ctx context.Context, _ interface{}) (string, string) {
		return c.subjType, internal.ValueOrEmpty(ctx, ctxKey)
	}

	return c
}

func (c *CheckMiddleware) WithObjectFromContextValue(ctxKey interface{}) *CheckMiddleware {
	c.objMapper = func(ctx context.Context, _ interface{}) (string, string) {
		return c.objType, internal.ValueOrEmpty(ctx, ctxKey)
	}

	return c
}

func (c *CheckMiddleware) WithPermissionMapper(permissionMapper PermissionMapper) *CheckMiddleware {
	c.permissionMapper = permissionMapper
	return c
}

func (c *CheckMiddleware) WithSubjectMapper(subjectMapper TypeIDMapper) *CheckMiddleware {
	c.subjMapper = subjectMapper
	return c
}

func (c *CheckMiddleware) WithObjectMapper(objectMapper TypeIDMapper) *CheckMiddleware {
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
		dsReader:       reader,
		ignoredMethods: []string{},
		subjType:       DefaultSubjType,
		objType:        DefaultObjType,
		ignoreCtx:      map[interface{}][]string{},
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

	objectType, objectID := c.object(ctx, req)
	if objectID == "" {
		return errors.New("object ID is empty")
	}

	subjectType, subjectID := c.subject(ctx, req)

	permission := ""
	if c.permissionMapper != nil {
		permission = c.permissionMapper(ctx, req)
	} else {
		permission = methodResource(ctx)
	}

	for _, path := range c.ignoredMethods {
		if strings.EqualFold(path, permission) {
			return nil
		}
	}

	allowed, err := c.dsReader.CheckPermission(ctx, &ds3.CheckPermissionRequest{
		SubjectType: subjectType,
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

func (c *CheckMiddleware) object(ctx context.Context, req interface{}) (string, string) {
	var objectType, objectID string

	if c.objMapper != nil {
		objectType, objectID = c.objMapper(ctx, req)
	}

	if c.objType == "" {
		objectType = c.objType
	}

	return objectType, objectID
}

func (c *CheckMiddleware) subject(ctx context.Context, req interface{}) (string, string) {
	var subjectType, subjectID string

	if c.subjMapper != nil {
		subjectType, subjectID = c.subjMapper(ctx, req)
	}

	if c.subjType == "" {
		subjectType = c.subjType
	}

	return subjectType, subjectID
}
