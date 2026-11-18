package models

import (
	"time"

	"gorm.io/gorm"
)

type Invitation struct {
	gorm.Model

	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
	RevokedAt *time.Time

	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}
