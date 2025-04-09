package grpcz

import (
	"context"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-aserto/middleware/internal"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	ds3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type (
	ObjectMapper func(ctx context.Context, req any) (objType, id string)
	Filter       func(ctx context.Context, req any) bool
)

type CheckClient interface {
	Check(ctx context.Context, in *ds3.CheckRequest, opts ...grpc.CallOption) (*ds3.CheckResponse, error)
}

type CheckOption func(*CheckOptions)

type objectSpecifier struct {
	id       string
	objType  string
	idMapper StringMapper
	mapper   ObjectMapper
}

func (os *objectSpecifier) resolve(ctx context.Context, req any) (string, string) {
	objType := os.objType
	objID := os.id

	switch {
	case os.mapper != nil:
		objType, objID = os.mapper(ctx, req)
	case os.idMapper != nil:
		objID = os.idMapper(ctx, req)
	}

	return objType, objID
}

// CheckOptions is used to configure the check middleware.
type CheckOptions struct {
	obj  objectSpecifier
	subj objectSpecifier
	rel  struct {
		name   string
		mapper StringMapper
	}
	filters []Filter
}

func (o *CheckOptions) object(ctx context.Context, req any) (string, string) {
	return o.obj.resolve(ctx, req)
}

func (o *CheckOptions) subject(ctx context.Context, req any) (string, string) {
	subjType, subjID := o.subj.resolve(ctx, req)
	if subjType == "" {
		subjType = internal.DefaultSubjType
	}

	return subjType, subjID
}

func (o *CheckOptions) relation(ctx context.Context, req any) string {
	relation := o.rel.name
	if o.rel.mapper != nil {
		relation = o.rel.mapper(ctx, req)
	}

	return relation
}

// WithRelation sets the relation/permission to check. If not specified, the relation is determined from the incoming request.
func WithRelation(name string) CheckOption {
	return func(o *CheckOptions) {
		o.rel.name = name
	}
}

// WithRelation takes a function that is used to determine the relation/permission to check from the incoming request.
func WithRelationMapper(mapper StringMapper) CheckOption {
	return func(o *CheckOptions) {
		o.rel.mapper = mapper
	}
}

// WithObjectType sets the object type to check.
func WithObjectType(objType string) CheckOption {
	return func(o *CheckOptions) {
		o.obj.objType = objType
	}
}

// WithObjectID set the id of the object to check.
func WithObjectID(id string) CheckOption {
	return func(o *CheckOptions) {
		o.obj.id = id
	}
}

// WithObjectIDFromContextValue takes the specified context value from the incoming request context and uses it as the object id to check.
func WithObjectIDFromContextValue(ctxKey any) CheckOption {
	return func(o *CheckOptions) {
		o.obj.idMapper = func(ctx context.Context, _ any) string {
			return internal.ValueOrEmpty(ctx, ctxKey)
		}
	}
}

// WithObjectIDMapper takes a function that is used to determine the object id to check from the incoming request.
func WithObjectIDMapper(mapper StringMapper) CheckOption {
	return func(o *CheckOptions) {
		o.obj.idMapper = mapper
	}
}

// WithObjectMapper takes a function that is used to determine the object type and id to check from the incoming request.
func WithObjectMapper(mapper ObjectMapper) CheckOption {
	return func(o *CheckOptions) {
		o.obj.mapper = mapper
	}
}

// WithSubjectType sets the subject type to check. Default is "user".
func WithSubjectType(subjType string) CheckOption {
	return func(o *CheckOptions) {
		o.subj.objType = subjType
	}
}

// WithSubjectID set the id of the subject to check.
func WithSubjectID(id string) CheckOption {
	return func(o *CheckOptions) {
		o.subj.id = id
	}
}

// WithSubjectIDFromContextValue takes the specified context value from the incoming request context and uses it as the
// subject id to check.
func WithSubjectIDFromContextValue(ctxKey any) CheckOption {
	return func(o *CheckOptions) {
		o.subj.idMapper = func(ctx context.Context, _ any) string {
			return internal.ValueOrEmpty(ctx, ctxKey)
		}
	}
}

// WithSubjectIDMapper takes a function that is used to determine the subject id to check from the incoming request.
func WithSubjectIDMapper(mapper StringMapper) CheckOption {
	return func(o *CheckOptions) {
		o.subj.idMapper = mapper
	}
}

// WithSubjectMapper takes a function that is used to determine the subject type and id to check from the incoming request.
func WithSubjectMapper(mapper ObjectMapper) CheckOption {
	return func(o *CheckOptions) {
		o.subj.mapper = mapper
	}
}

func WithMethodFilter(methods ...string) CheckOption {
	lookup := internal.NewLookup(methods...)

	return func(o *CheckOptions) {
		o.filters = append(o.filters, func(ctx context.Context, _ any) bool {
			method, _ := grpc.Method(ctx)
			return lookup.Contains(method)
		})
	}
}

func WithContextValueFilter(ctxKey any, values ...string) CheckOption {
	lookup := internal.NewLookup(values...)

	return func(o *CheckOptions) {
		o.filters = append(o.filters, func(ctx context.Context, _ any) bool {
			value := internal.ValueOrEmpty(ctx, ctxKey)
			return lookup.Contains(value)
		})
	}
}

func WithFilter(filter Filter) CheckOption {
	return func(o *CheckOptions) {
		o.filters = append(o.filters, filter)
	}
}

type CheckMiddleware struct {
	dsClient CheckClient
	opts     *CheckOptions
}

func NewCheckMiddleware(client CheckClient, options ...CheckOption) *CheckMiddleware {
	opts := &CheckOptions{}
	for _, o := range options {
		o(opts)
	}

	if opts.rel.name == "" && opts.rel.mapper == nil {
		opts.rel.mapper = relationFromMethod
	}

	return &CheckMiddleware{
		dsClient: client,
		opts:     opts,
	}
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := c.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
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

func (c *CheckMiddleware) authorize(ctx context.Context, req any) error {
	for _, filter := range c.opts.filters {
		if filter(ctx, req) {
			return nil
		}
	}

	objType, objID := c.opts.object(ctx, req)
	if objID == "" {
		return errors.New("object ID is empty")
	}

	if objType == "" {
		return errors.New("object type is empty")
	}

	subjType, subjID := c.opts.subject(ctx, req)
	if subjID == "" {
		return errors.New("subject ID is empty")
	}

	if subjType == "" {
		return errors.New("subject type is empty")
	}

	check := &ds3.CheckRequest{
		ObjectType:  objType,
		ObjectId:    objID,
		Relation:    c.opts.relation(ctx, req),
		SubjectType: subjType,
		SubjectId:   subjID,
	}

	logger := zerolog.Ctx(ctx).With().Interface("check_request", check).Logger()
	logger.Debug().Msg("authorizing request")
	ctx = logger.WithContext(ctx)

	allowed, err := c.dsClient.Check(ctx, check)
	if err != nil {
		return cerr.WrapContext(err, ctx, "check call failed")
	}

	if !allowed.GetCheck() {
		return cerr.WithContext(aerr.ErrAuthorizationFailed, ctx)
	}

	return nil
}

func relationFromMethod(ctx context.Context, _ any) string {
	return permissionFromMethod(ctx)
}
