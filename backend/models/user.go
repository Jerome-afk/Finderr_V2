package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`

	IsActive   bool `gorm:"default:true"`
	IsVerified bool `gorm:"not null;default:false"`

	// API Keys
	AnilistID string `gorm:"unique"`

	Roles []Role `gorm:"many2many:user_roles;"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
