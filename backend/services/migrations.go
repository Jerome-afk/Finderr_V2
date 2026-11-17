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
	)
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}
