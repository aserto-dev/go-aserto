package grpcz

import (
	"context"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func (b *IdentityBuilder) InternalBuild(ctx context.Context, req any) *api.IdentityContext {
	return b.build(ctx, req)
}
