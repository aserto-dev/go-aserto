package httpz_test

import (
	"log"
	"net/http"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-aserto/middleware/httpz"
)

const readHeaderTimeout = 2 * time.Second

func Hello(w http.ResponseWriter, _ *http.Request) {
	if _, err := w.Write([]byte(`"hello"`)); err != nil {
		log.Println("Failed to write HTTP response:", err)
	}
}

func Example() {
	// Create azClient client.
	azClient, err := az.New(
		aserto.WithAPIKeyAuth("<Aserto authorizer API Key>"),
		aserto.WithTenantID("<Aserto tenant ID>"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer client:", err)
	}

	defer func() { _ = azClient.Close() }()

	// Create HTTP middleware.
	middleware := httpz.New(
		azClient,
		&httpz.Policy{
			Name:     "<Aserto policy Name>",
			Decision: "<authorization decision (e.g. 'allowed')",
		},
	)

	// Define HTTP route.
	http.Handle(
		"/",
		middleware.Handler(http.HandlerFunc(Hello)), // Attach middleware to route.
	)

	// Start server.
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: readHeaderTimeout,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Println("Failed to start server:", err)
	}
}
