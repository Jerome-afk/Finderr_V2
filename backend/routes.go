package main

import (
	"finderr/handlers"
	"finderr/sessions"
	"finderr/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Get database connection pool
	sqlDB, err := utils.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Initialise session manager
	store := sessions.NewPostgreSQLStore(sqlDB)

	sessions, err := sessions.NewSessionManager(
		store,
		sessions.SessionConfig{
			IdleExpiration:     30 * time.Minute,
			AbsoluteExpiration: 24 * time.Hour,
			GCInterval:         10 * time.Minute,
			CookieName:         "media_session",
			SecureCookie:       true,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer sessions.Close()

	// Use session middleware
	r.Use(sessions.Middleware())

	// Load HTML templates
	r.LoadHTMLGlob("templates/**/*")

	// Serve static files
	r.Static("/static", "./static")

	// Define routes
	auth := r.Group("/")

	auth.GET("/", handlers.Home)

	return r
}
