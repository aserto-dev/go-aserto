package internal

import (
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func DefaultPolicyContext(policy *middleware.Policy) *api.PolicyContext {
	return &api.PolicyContext{
		Path:      policy.Path,
		Decisions: []string{policy.Decision},
	}
}
