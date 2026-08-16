package services

import (
	"finderr/models"
	"finderr/utils"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func MigrateDB() {
	err := utils.DB.AutoMigrate(&models.User{})

	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}

func CreateAdmin() {
	var admin models.User

	// Check if the admin user already exists
	result := utils.DB.Where("role = ?", "admin").First(&admin)

	if result.Error == nil {
		// Admin user already exists, no need to create
		log.Println("Admin user already exists.")
		return
	}

	// Create a new admin user
	// Hash the pass
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password: ", err)
	}

	admin = models.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := utils.DB.Create(&admin).Error; err != nil {
		log.Fatal("Failed to create admin user: ", err)
	}

	log.Println("Admin user created successfully.")
}
