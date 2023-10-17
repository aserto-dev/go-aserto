package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/aserto-dev/go-aserto/middleware/grpc"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

const username = "george"

func Anon() *api.IdentityContext {
	return &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE}
}

func JWT() *api.IdentityContext {
	return &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_JWT, Identity: username}
}

func SUB() *api.IdentityContext {
	return &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_SUB, Identity: username}
}

func TestTypeAssignment(t *testing.T) {
	assert.Equal(
		t,
		JWT(),
		(&grpc.IdentityBuilder{}).JWT().ID(username).InternalBuild(context.TODO(), nil),
		"Expected JWT identity type",
	)
}

func TestAssignmentOverride(t *testing.T) {
	identity := (&grpc.IdentityBuilder{}).JWT().None().InternalBuild(context.TODO(), nil)

	assert.Equal(
		t,
		Anon(),
		(&grpc.IdentityBuilder{}).JWT().None().InternalBuild(context.TODO(), nil),
		identity.Type,
		"Expected NONE identity to override JWT",
	)
}

func TestAssignmentOrder(t *testing.T) {
	assert.Equal(
		t,
		(&grpc.IdentityBuilder{}).ID("id").Subject(),
		(&grpc.IdentityBuilder{}).Subject().ID("id"),
		"Assignment order shouldn't matter",
	)
}

func TestNoneClearsIdentity(t *testing.T) {
	assert.Equal(
		t,
		Anon(),
		(&grpc.IdentityBuilder{}).ID("id").None().InternalBuild(context.TODO(), nil),
		"WithNone should override previously assigned identity",
	)
}

func TestIdentityFromMetadata(t *testing.T) {
	builder := &grpc.IdentityBuilder{}
	builder.JWT().FromMetadata("authorization")

	md := metadata.New(map[string]string{"authorization": username})
	ctx := metadata.NewIncomingContext(context.TODO(), md)

	assert.Equal(
		t,
		JWT(),
		builder.InternalBuild(ctx, nil),
		"Identity should be read from context metadata",
	)
}

func TestIdentityFromMissingMetadata(t *testing.T) {
	builder := &grpc.IdentityBuilder{}
	builder.JWT().FromMetadata("authorization")

	md := metadata.New(map[string]string{"wrongKey": username})
	ctx := metadata.NewIncomingContext(context.TODO(), md)

	assert.Equal(
		t,
		Anon(),
		builder.InternalBuild(ctx, nil),
		"Missing metadata value results in anonymous identity",
	)
}

func TestIdentityFromMissingMetadataValue(t *testing.T) {
	builder := &grpc.IdentityBuilder{}
	builder.JWT().FromMetadata("authorization")

	assert.Equal(
		t,
		Anon(),
		builder.InternalBuild(context.TODO(), nil),
		"Missing metadata results in anonymous identity",
	)
}

type user struct{}

func TestIdentityFromContextValue(t *testing.T) {
	builder := &grpc.IdentityBuilder{}
	builder.Subject().FromContextValue(user{})

	ctx := context.WithValue(context.TODO(), user{}, "george")

	assert.Equal(
		t,
		SUB(),
		builder.InternalBuild(ctx, nil),
		"Identity should be read from context value",
	)
}

func TestMissingContextValue(t *testing.T) {
	builder := &grpc.IdentityBuilder{}
	builder.Subject().FromContextValue(user{})

	assert.Equal(
		t,
		Anon(),
		builder.InternalBuild(context.TODO(), nil),
		"Missing context value should result in anonymous identity",
	)
}
