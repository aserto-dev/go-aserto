package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/humaz"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

const port = 8080
const contextKey = "subject"
const subjectValue = "CiRmZDE2MTRkMy1jMzlhLTQ3ODEtYjdiZC04Yjk2ZjVhNTEwMGQSBWxvY2Fs"

func AuthNMiddleware(ctx huma.Context, next func(huma.Context)) {
	ctx = huma.WithValue(ctx, contextKey, subjectValue)
	next(ctx)
}

func main() {
	// Create Aserto authorizer client
	azClient, err := az.New(
		aserto.WithAddr("localhost:8282"),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}
	defer azClient.Close()

	// Create Aserto middleware for Huma
	mw := humaz.New(
		azClient,
		&middleware.Policy{
			Name:     "local",
			Decision: "allowed",
		},
	)

	mw.Identity.FromContextValue(contextKey)
	mw.Identity.Manual()

	// Set up Gin router
	router := gin.Default()

	// Initialize Huma API with Gin adapter
	api := humagin.New(router, huma.DefaultConfig("Aserto Example", "1.0.0"))
	// Configure authorization middlewares for all operations
	api.UseMiddleware(AuthNMiddleware, mw.Handler)

	huma.Register(api, huma.Operation{
		OperationID: "getAsset",
		Method:      "GET",
		Path:        "/api/{asset}",
		Summary:     "Get an asset",
		// Configure authorization only on per operation basis
		// Middlewares: huma.Middlewares{mw.Handler},
	}, handler)

	huma.Register(api, huma.Operation{
		OperationID: "createAsset",
		Method:      "POST",
		Path:        "/api/{asset}",
		Summary:     "Create an asset",
	}, handler)

	huma.Register(api, huma.Operation{
		OperationID: "deleteAsset",
		Method:      "DELETE",
		Path:        "/api/{asset}",
		Summary:     "Delete an asset",
	}, handler)

	// Start the server
	fmt.Printf("Server running on port %d\n", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal(err)
	}
}

// Input struct for the handler
type AssetRequest struct {
	Asset string `path:"asset"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, input *AssetRequest) (*SuccessResponse, error) {
	return &SuccessResponse{Message: "Permission granted"}, nil
}
