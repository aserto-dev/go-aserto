package gorillaz_test

import (
	"log"
	"net/http"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
	mw "github.com/aserto-dev/go-aserto/middleware/gorillaz"
)

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
	defer azClient.Close()

	// Create HTTP middleware.
	middleware := mw.New(
		azClient,
		&mw.Policy{
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
		ReadHeaderTimeout: 2 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Println("Failed to start server:", err)
	}
}
