package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model

	ID      uint    `gorm:"primaryKey"`
	Name    string   `gorm:"size:100;uniqueIndex;not null"`
	Description string   `gorm:"size:255"`

	CreatedAt time.Time
	UpdatedAt time.Time
}