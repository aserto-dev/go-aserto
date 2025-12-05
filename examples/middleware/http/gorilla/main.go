package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/gorillaz"
)

const (
	port              = 8080
	readHeaderTimeout = 2 * time.Second
)

func main() {
	azClient, err := az.New(
		aserto.WithAddr("localhost:8282"),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}

	defer func() { _ = azClient.Close() }()

	mw := gorillaz.New(
		azClient,
		&middleware.Policy{
			Name:     "local",
			Decision: "allowed",
		},
	)
	mw.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
		if username, _, ok := r.BasicAuth(); ok {
			identity.Subject().ID(username)
		}
	})
	mw.WithPolicyFromURL("example")

	router := mux.NewRouter()
	router.HandleFunc("/api/{asset}", Handler).Methods("GET", "POST", "DELETE")

	router.Use(mw.Handler)
	start(router)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	if _, err := w.Write([]byte(`"Permission granted"`)); err != nil {
		log.Fatal(err)
	}
}

func start(h http.Handler) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("Staring server on %s\n", addr)

	srv := http.Server{
		Handler:           h,
		Addr:              addr,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
