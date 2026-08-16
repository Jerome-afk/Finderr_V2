package main

import (
	"finderr/services"
	"finderr/utils"
)

func main() {
	// Connect to the database
	utils.ConnectDB()

	// Run migrations
	services.MigrateDB()

	// Create admin user if not exists
	services.CreateAdmin()

	// Setup routes and start the server
	router := SetupRoutes()

	router.Run(":8080")
}
