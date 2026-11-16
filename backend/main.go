package main

import (
	"finderr/services"
	"finderr/utils"
)

func main() {
	// Initialize the logger
	logger, err := utils.NewLogger("logs", utils.INFO)
	if err != nil {
		panic(err)
	}
	// Connect to the database
	logger.Info("Establishing connection to database...")
	utils.ConnectDB()
	logger.Info("Connected to the database")

	// Run migrations
	services.MigrateDB()

	// Create admin user if not exists
	services.CreateAdmin()

	// Setup routes and start the server
	router := SetupRoutes()

	router.Run(":8080")
}
