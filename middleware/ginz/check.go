package ginz

import (
	"fmt"
	"net/http"

	"github.com/aserto-dev/go-aserto/middleware/internal"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/structpb"
)

// CheckOption is used to configure the check middleware.
type CheckOption func(*CheckOptions)

// ObjectMapper takes an incoming request and returns the object type and id to check.
type ObjectMapper func(*gin.Context) (objType string, id string)

// WithIdentityMapper takes an identity mapper function that is used to determine the subject id for the check call.
func WithIdentityMapper(mapper IdentityMapper) CheckOption {
	return func(o *CheckOptions) {
		o.subj.mapper = mapper
	}
}

// WithRelation sets the relation/permission to check.
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

// WithObjectIDMapper takes a function that is used to determine the object id to check from the incoming request.
func WithObjectIDMapper(mapper StringMapper) CheckOption {
	return func(o *CheckOptions) {
		o.obj.idMapper = mapper
	}
}

// WithObjectIDFromVar takes the name of a variable in the request path that is used as the object id to check.
func WithObjectIDFromVar(name string) CheckOption {
	return func(o *CheckOptions) {
		o.obj.idMapper = func(g *gin.Context) string {
			return g.Param(name)
		}
	}
}

// WithObjectMapper takes a function that is used to determine the object type and id to check from the incoming request.
func WithObjectMapper(mapper ObjectMapper) CheckOption {
	return func(o *CheckOptions) {
		o.obj.mapper = mapper
	}
}

// WithPolicyPath sets the path of the policy module to use for the check call.
func WithPolicyPath(path string) CheckOption {
	return func(o *CheckOptions) {
		o.policy.path = path
	}
}

// CheckOptions is used to configure the check middleware.
type CheckOptions struct {
	obj struct {
		id       string
		objType  string
		idMapper StringMapper
		mapper   ObjectMapper
	}
	rel struct {
		name   string
		mapper StringMapper
	}
	subj struct {
		subjType string
		mapper   IdentityMapper
	}
	policy struct {
		path   string
		mapper StringMapper
	}
}

func (o *CheckOptions) object(g *gin.Context) (string, string) {
	objType := o.obj.objType
	objID := o.obj.id

	switch {
	case o.obj.mapper != nil:
		objType, objID = o.obj.mapper(g)
	case o.obj.idMapper != nil:
		objID = o.obj.idMapper(g)
	}

	return objType, objID
}

func (o *CheckOptions) relation(g *gin.Context) string {
	relation := o.rel.name
	if o.rel.mapper != nil {
		relation = o.rel.mapper(g)
	}

	return relation
}

func (o *CheckOptions) subjectType() string {
	if o.subj.subjType != "" {
		return o.subj.subjType
	}

	return internal.DefaultSubjType
}

type Check struct {
	mw   *Middleware
	opts *CheckOptions
}

func newCheck(mw *Middleware, options ...CheckOption) *Check {
	opts := &CheckOptions{}
	for _, o := range options {
		o(opts)
	}

	return &Check{mw: mw, opts: opts}
}

// Handler returns a middleware handler that checks incoming requests.
func (c *Check) Handler(g *gin.Context) {
	policyContext := c.policyContext(g)
	identityContext := c.identityContext(g)

	resourceContext, err := c.resourceContext(g)
	if err != nil {
		_ = g.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	allowed, err := c.mw.is(g.Request.Context(), identityContext, policyContext, resourceContext)
	if err != nil {
		_ = g.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !allowed {
		g.AbortWithStatus(http.StatusForbidden)
		return
	}

	g.Next()
}

func (c *Check) policyContext(g *gin.Context) *api.PolicyContext {
	policyContext := c.mw.policyContext()
	policyContext.Path = ""

	if c.opts.policy.path != "" {
		policyContext.Path = c.opts.policy.path
	}

	policyMapper := c.opts.policy.mapper
	if policyMapper != nil {
		policyContext.Path = policyMapper(g)
	}

	if policyContext.GetPath() == "" {
		path := "check"
		if c.mw.policy.Root != "" {
			path = fmt.Sprintf("%s.%s", c.mw.policy.Root, path)
		}

		policyContext.Path = path
	}

	return policyContext
}

func (c *Check) identityContext(g *gin.Context) *api.IdentityContext {
	idc := c.mw.Identity.Build(g)

	if c.opts.subj.mapper != nil {
		identity := internal.NewIdentity(idc.GetType(), idc.GetIdentity())
		c.opts.subj.mapper(g, identity)
		idc = identity.Context()
	}

	return idc
}

func (c *Check) resourceContext(g *gin.Context) (*structpb.Struct, error) {
	relation := c.opts.relation(g)
	objType, objID := c.opts.object(g)
	subjType := c.opts.subjectType()

	return structpb.NewStruct(map[string]any{
		"relation":     relation,
		"object_type":  objType,
		"object_id":    objID,
		"subject_type": subjType,
	})
}
