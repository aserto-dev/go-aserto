package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/authorizer"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/http/ginz"
	"github.com/gin-gonic/gin"
)

const port = 8080

func main() {
	azClient, err := authorizer.New(
		aserto.WithAddr("localhost:8282"),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}
	defer azClient.Close()

	mw := ginz.New(
		azClient.Authorizer,
		&middleware.Policy{
			Name:          "local",
			Decision:      "allowed",
			InstanceLabel: "label",
		},
	)
	mw.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
		if username, _, ok := r.BasicAuth(); ok {
			identity.Subject().ID(username)
		}
	})
	mw.WithPolicyFromURL("example")

	router := gin.Default()
	router.Use(mw.Handler)
	router.GET("/api/:asset", Handler)
	router.POST("/api/:asset", Handler)
	router.DELETE("/api/:asset", Handler)

	router.Run(fmt.Sprintf(":%d", port))
}

func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, "Permission granted")
}
