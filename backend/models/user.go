package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"not null;default:'user'"` // Default role is 'user'
	AnilistID string `gorm:"unique"`         // Unique Anilist ID for each user
	
}
