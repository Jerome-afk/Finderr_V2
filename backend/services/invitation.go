package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"finderr/models"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const invitationLifetime = 24 * time.Hour

type CreateInvitatiionInput struct {
	Username     string
	Email        string
	RoleNames    []string
}

type CreatedInvitation struct {
	UserID   uint
	Email    string
	Token    string
	ExpiresAt time.Time
}

func CreateInvitation(
	db *gorm.DB,
	actorID uint,
	input CreateInvitatiionInput,
) (*CreatedInvitation, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Username == "" || len([]rune(input.Username)) > 150 {
		return nil, errors.New("name must contain between 1 to 150 characters")
	}

	if len(input.Email) > 255 {
		return nil, errors.New("email is too long")
	}

	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email {
		return nil, errors.New("invalid email address")
	}

	roleNames := make([]string, 0, len(input.RoleNames))
	seen := make(map[string]bool)

	for _, name := range input.RoleNames {
		name = strings.ToLower(strings.TrimSpace(name))

		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		roleNames = append(roleNames, name)
	}

	if len(roleNames) == 0 {
		return nil, errors.New("at least one role must be assigned")
	}

	// Generate a secure, random invitation token
	randomToken := make([]byte, 32)

	if _, err := rand.Read(randomToken); err != nil {
		return nil, fmt.Errorf("generate invitation token: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomToken)

	tokenHashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	expiresAt := time.Now().UTC().Add(invitationLifetime)

	var createdUser models.User

	err = db.Transaction(func(tx *gorm.DB) error {
		// Load administrator and their respective permissions
		var actor models.User

		if err := tx.Preload("Roles.Permissions").
			First(&actor, actorID).Error; err != nil {
				return fmt.Errorf("load administrator: %w", err)
		}

		if !actor.IsActive {
			return errors.New("administrator account is inactive")
		}

		isSuperAdmin := false
		actorPermissions := make(map[string]bool)

		for _, role := range actor.Roles {
			if role.Name == "super_admin" {
				isSuperAdmin = true
			}

			for _, permission := range role.Permissions {
				actorPermissions[permission.Name] = true
			}
		}

		if !isSuperAdmin && !actorPermissions["users.create"] {
			return errors.New("you do not have permission to create users")
		}

		if !isSuperAdmin && !actorPermissions["roles.assign"] {
			return errors.New("you do not have permission to assign roles")
		}

		// Find all roles requested
		var roles []models.Role

		if err := tx.Preload("Permissions").
			Where("name IN ?", roleNames).
			Find(&roles).Error; err != nil {
				return fmt.Errorf("load requested roles: %w", err)
		}

		if len(roles) != len(roleNames) {
			return errors.New("one or more requested roles do not exist")
		}

		for _, role := range roles {
			if role.Name == "super_admin" && !isSuperAdmin {
				return errors.New("only a Super Admin can assign the super_admin role")
			}

			if !isSuperAdmin {
				for _, permission := range role.Permissions {
					if !actorPermissions[permission.Name] {
						return fmt.Errorf(
							"you cannot assign role %q because you lack permission %q",
							role.Name,
							permission.Name,
						)
					}
				}
			}
		}

		// Prevent reuse of email
		var existing int64

		if err := tx.Unscoped().
			Model(&models.User{}).
			Where("LOWER(email) = ?", input.Email).
			Count(&existing).Error; err != nil {
				return fmt.Errorf("check email: %w", err)
		}

		if existing > 0 {
			return errors.New("an account with this email already exists")
		}

		// Create the account in an inactive state
		createdUser = models.User{
			Username: input.Username,
			Email: input.Email,
			PasswordHash: "",
			IsActive: false,
			IsVerified: false,
		}

		if err := tx.Create(&createdUser).Error; err != nil {
			return fmt.Errorf("created user: %w", err)
		}

		if err := tx.Model(&createdUser).
			Association("Roles").
			Replace(&roles); err != nil {
				return fmt.Errorf("assign roles: %w", err)
		}

		invitation := models.Invitation{
			UserID: createdUser.ID,
			TokenHash: tokenHash,
			ExpiresAt: expiresAt,
		}

		if err := tx.Create(&invitation).Error; err != nil {
			return fmt.Errorf("create invitation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreatedInvitation{
		UserID: createdUser.ID,
		Email: createdUser.Email,
		Token: rawToken,
		ExpiresAt: expiresAt,
	}, nil
}

func HashInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ValidateTokenFormat(token string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(decoded) == 32
}

var _ = clause.Locking{}