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
	_ context.Context,
	_ *authz.DecisionTreeRequest,
	_ ...grpc.CallOption,
) (*authz.DecisionTreeResponse, error) {
	return nil, nil
}

func (c *Authorizer) Is(
	_ context.Context,
	in *authz.IsRequest,
	_ ...grpc.CallOption,
) (*authz.IsResponse, error) {
	assert.Equal(c.t, c.expected, in)
	return &c.response, nil
}

func (c *Authorizer) Query(
	_ context.Context,
	_ *authz.QueryRequest,
	_ ...grpc.CallOption,
) (*authz.QueryResponse, error) {
	return nil, nil
}

func (c *Authorizer) Compile(
	_ context.Context,
	_ *authz.CompileRequest,
	_ ...grpc.CallOption,
) (*authz.CompileResponse, error) {
	return nil, nil
}

func (c *Authorizer) GetPolicy(
	_ context.Context,
	_ *authz.GetPolicyRequest,
	_ ...grpc.CallOption,
) (*authz.GetPolicyResponse, error) {
	return nil, nil
}

func (c *Authorizer) ListPolicies(
	_ context.Context,
	_ *authz.ListPoliciesRequest,
	_ ...grpc.CallOption,
) (*authz.ListPoliciesResponse, error) {
	return nil, nil
}

func (c *Authorizer) Info(
	_ context.Context,
	_ *authz.InfoRequest,
	_ ...grpc.CallOption,
) (*authz.InfoResponse, error) {
	return nil, nil
}
