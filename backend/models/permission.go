package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model

	ID      uint    `gorm:"primaryKey"`
	Username    string   `gorm:"size:100;uniqueIndex;not null"`
	Description string   `gorm:"size:255"`
}