package services

import (
	"errors"
	"finderr/auth"
	"finderr/models"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		return fmt.Errorf("check invitation: %w", err)
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Lock the invitation to prevent double activation
		var lockedInvitation models.Invitation

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"token_hash = ? AND used_at IS NULL AND revoked_at IS NULL AND expires_at > ?",
				tokenHash,
				now,
			).
			First(&lockedInvitation).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired invitation")
			}
			return fmt.Errorf("lock invitation: %w", err)
		}

		var user models.User

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&user, lockedInvitation.UserID).Error; err != nil {
				return fmt.Errorf("load invited user: %w", err)
		}

		if user.IsActive {
			return errors.New("this account has already been activated")
		}

		//TODO: Activate the account

	})

	return nil
}