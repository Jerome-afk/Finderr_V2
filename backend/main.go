package main

import (
	"finderr/services"
	"finderr/utils"
	"log"
)

func main() {
	// Initialize the logger
	logger, err := utils.NewLogger("logs", utils.INFO)
	if err != nil {
		panic(err)
	}
	// Connect to the database
	logger.Info("Establishing connection to database...")
	if err := utils.ConnectDB(); err != nil {
		log.Fatal(err)
	}
	logger.Info("Connected to the database")

	// Run migrations
	services.MigrateDB()

	// Setup routes and start the server
	router := SetupRoutes()

	router.Run(":8080")
}
