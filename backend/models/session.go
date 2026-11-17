package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model

	ID           string    `gorm:"primaryKey;type:varchar(64)"`
	UserID        uint      `gorm:"not null;index"`
	Data         []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastActivity time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
}

func (Session) TableName() string {
	return "sessions"
}
