package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/ginz"
	"github.com/gin-gonic/gin"
)

const port = 8080

func main() {
	azClient, err := az.New(
		aserto.WithAddr("localhost:8282"),
	)
	if err != nil {
		log.Fatalln("Failed to create authorizer client:", err)
	}

	defer func() { _ = azClient.Close() }()

	mw := ginz.New(
		azClient,
		&middleware.Policy{
			Name:     "local",
			Decision: "allowed",
		},
	)
	mw.Identity.Mapper(func(c *gin.Context, identity middleware.Identity) {
		if username, _, ok := c.Request.BasicAuth(); ok {
			identity.Subject().ID(username)
		}
	})
	mw.WithPolicyFromURL("example")

	router := gin.Default()
	router.Use(mw.Handler)
	router.GET("/api/:asset", Handler)
	router.POST("/api/:asset", Handler)
	router.DELETE("/api/:asset", Handler)

	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Print(err)
		return
	}
}

func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, "Permission granted")
}
