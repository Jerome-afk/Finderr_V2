package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"size:255;uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	IsActive  bool   `gorm:"default:true"`
	IsVerified bool `gorm:"not null;default:true"`
	AnilistID string `gorm:"unique"`         // Unique Anilist ID for each user
	CreatedAt time.Time
	UpdatedAt time.Time

	Roles []Role `gorm:"many2many:user_roles;"`
}
