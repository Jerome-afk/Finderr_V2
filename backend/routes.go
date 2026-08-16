package main

import (
	"finderr/handlers"
	"finderr/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Session store setup
	store := cookie.NewStore([]byte("bettysupersecret"))

	// Use sessions middleware
	r.Use(sessions.Sessions("mysession", store))

	// Load HTML templates
	r.LoadHTMLGlob("templates/**/*")

	// Serve static files
	r.Static("/static", "./static")

	// Define routes
	r.GET("/login", handlers.LoginPage)
	r.POST("/login", handlers.Login)

	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())

	auth.GET("/", handlers.Home)

	return r
}
