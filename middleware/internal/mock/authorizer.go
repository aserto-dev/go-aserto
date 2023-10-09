package mock

import (
	"context"
	"testing"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type Authorizer struct {
	t        *testing.T
	expected *authz.IsRequest
	response authz.IsResponse
}

func New(t *testing.T, expectedRequest *authz.IsRequest, decision *authz.Decision) *Authorizer {
	return &Authorizer{
		t:        t,
		expected: expectedRequest,
		response: authz.IsResponse{
			Decisions: []*authz.Decision{decision},
		},
	}
}

var _ authz.AuthorizerClient = (*Authorizer)(nil)

func (c *Authorizer) DecisionTree(
	ctx context.Context,
	in *authz.DecisionTreeRequest,
	opts ...grpc.CallOption,
) (*authz.DecisionTreeResponse, error) {
	return nil, nil
}

func (c *Authorizer) Is(
	ctx context.Context,
	in *authz.IsRequest,
	opts ...grpc.CallOption,
) (*authz.IsResponse, error) {
	assert.Equal(c.t, c.expected, in)
	return &c.response, nil
}

func (c *Authorizer) Query(
	ctx context.Context,
	in *authz.QueryRequest,
	opts ...grpc.CallOption,
) (*authz.QueryResponse, error) {
	return nil, nil
}

func (c *Authorizer) Compile(
	ctx context.Context,
	in *authz.CompileRequest,
	opts ...grpc.CallOption,
) (*authz.CompileResponse, error) {
	return nil, nil
}

func (c *Authorizer) GetPolicy(
	ctx context.Context,
	in *authz.GetPolicyRequest,
	opts ...grpc.CallOption,
) (*authz.GetPolicyResponse, error) {
	return nil, nil
}

func (c *Authorizer) ListPolicies(
	ctx context.Context,
	in *authz.ListPoliciesRequest,
	opts ...grpc.CallOption,
) (*authz.ListPoliciesResponse, error) {
	return nil, nil
}

func (c *Authorizer) Info(
	ctx context.Context,
	in *authz.InfoRequest,
	opts ...grpc.CallOption,
) (*authz.InfoResponse, error) {
	return nil, nil
}
