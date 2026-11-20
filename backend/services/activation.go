package services

import (
	"errors"
	"finderr/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ActivationAccountInput struct {
	Token    string
	Password string
}

func ActivationAccount(db *gorm.DB, input ActivationAccountInput) error {
	if db == nil {
		return errors.New("database connection is nil")
	}

	input.Token = strings.TrimSpace(input.Token)

	if !ValidateTokenFormat(input.Token) {
		return errors.New("invalid or expired invitation")
	}

	if len(input.Password) < 12 || len(input.Password) > 128 {
		return errors.New("password must be between 12 and 128 bytes")
	}

	tokenHash := HashInvitationToken(input.Token)
	now := time.Now().UTC()

	// Check that the invitation token exists
	var invitation models.Invitation

	err := db.Where(
		"token_hash = ? AND used_at IS NULL AND revoked_at IS NULL AND expires_at > ?",
		tokenHash,
		now,
	).First(&invitation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired invitation")
		}
	}
}