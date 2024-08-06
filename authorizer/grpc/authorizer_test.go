package grpc_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aserto-dev/go-aserto/authorizer/grpc"
	"github.com/aserto-dev/go-aserto/client"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func Example() {
	// Create new authorizer client.
	authorizer, err := grpc.New(
		client.WithAPIKeyAuth("<Aserto authorizer API key"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer:", err)
	}

	// Make an authorization call.
	result, err := authorizer.Is(
		context.Background(),
		&authz.IsRequest{
			PolicyContext: &api.PolicyContext{
				Path:      "<Policy path (e.g. 'peoplefinder.GET.users')",
				Decisions: []string{"<authorization decisions (e.g. 'allowed')>"},
			},
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
				Identity: "<user id>",
			},
			PolicyInstance: &api.PolicyInstance{
				Name:          "<Aserto Policy Name>",
				InstanceLabel: "<Aserto Policy Instance Label>",
			},
		},
	)
	if err != nil {
		log.Fatal("Failed to make authorization call:", err)
	}

	// Check the authorizer's decision.
	for _, decision := range result.Decisions {
		if decision.Decision == "allowed" { // "allowed" is just an example. Your policy may have different rules.
			if decision.Is {
				fmt.Println("Access granted")
			} else {
				fmt.Println("Access denied")
			}
		}
	}
}
