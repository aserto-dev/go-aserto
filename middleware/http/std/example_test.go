package std_test

import (
	"log"
	"net/http"
	"time"

	"github.com/aserto-dev/go-aserto/authorizer/grpc"
	"github.com/aserto-dev/go-aserto/client"
	mw "github.com/aserto-dev/go-aserto/middleware/http/std"
)

func Hello(w http.ResponseWriter, _ *http.Request) {
	if _, err := w.Write([]byte(`"hello"`)); err != nil {
		log.Println("Failed to write HTTP response:", err)
	}
}

func Example() {
	// Create authorizer client.
	authorizer, err := grpc.New(
		client.WithAPIKeyAuth("<Aserto authorizer API Key>"),
		client.WithTenantID("<Aserto tenant ID>"),
	)
	if err != nil {
		log.Fatal("Failed to create authorizer client:", err)
	}

	// Create HTTP middleware.
	middleware := mw.New(
		authorizer,
		&mw.Policy{
			Name:          "<Aserto policy Name>",
			Decision:      "<authorization decision (e.g. 'allowed')",
			InstanceLabel: "<Aserto  policy instance label>",
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
	log.Fatal(server.ListenAndServe())
}
