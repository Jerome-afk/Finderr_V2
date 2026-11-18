package services

import (
	"finderr/models"
	"finderr/utils"
)

func MigrateDB() {
	err := utils.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Session{},
		&models.Invitation{},
	)
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	if err := utils.SeedRoles(utils.DB); err != nil {
		panic("role seeding failed: " + err.Error())
	}
}
