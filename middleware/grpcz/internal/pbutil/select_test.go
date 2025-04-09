package pbutil_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/go-aserto/middleware/grpcz/internal/pbutil"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestFieldMaskIsValid(t *testing.T) {
	msg := &authz.IsRequest{
		PolicyContext: &api.PolicyContext{
			Path: "policy.path",
		},
		IdentityContext: &api.IdentityContext{
			Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			Identity: "username",
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "policyName",
			InstanceLabel: "label",
		},
	}

	var msgType *authz.IsRequest

	mask, err := fieldmaskpb.New(
		msgType,
		"policy_context.path",
		"identity_context.identity",
		"resource_context",
		"policy_instance.name",
		"policy_instance.instance_label",
	)

	require.NoError(t, err, "failed to create field mask")
	assert.True(t, mask.IsValid(msg), "invalid mask")

	mask.Normalize()

	testCase := func(paths []string, expected map[string]any) func(t *testing.T) {
		return func(t *testing.T) {
			selection, err := pbutil.Select(msg, paths...)
			require.NoError(t, err, "select failed on policy_context.path")

			actual := selection.AsMap()

			assert.True(t, reflect.DeepEqual(expected, actual), "wrong selection")
		}
	}

	t.Run("single value", testCase(
		[]string{"policy_context.path"},
		map[string]any{
			"policy_context": map[string]any{
				"path": "policy.path",
			},
		},
	))
	t.Run("multiple values", testCase(
		[]string{"policy_context.path", "identity_context.identity"},
		map[string]any{
			"policy_context": map[string]any{
				"path": "policy.path",
			},
			"identity_context": map[string]any{
				"identity": "username",
			},
		},
	))
	t.Run("struct value", testCase(
		[]string{"policy_context"},
		map[string]any{
			"policy_context": map[string]any{
				"path": "policy.path",
			},
		},
	))
}
