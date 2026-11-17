package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	ID          uint   `gorm:"primaryKey"`
	Username    string `gorm:"size:50;uniqueIndex;not null"`
	Description string `gorm:"size:255"`

	IsSystem    bool   `gorm:"not null;default:false"` // Prevent accidental modification

	Permissions []Permission `gorm:"many2many:role_permissions;"`

	CreatedAt   time.Time
	UpdatedAt   time.Time
}