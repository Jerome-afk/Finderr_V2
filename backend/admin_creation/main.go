package main

import (
	"bufio"
	"errors"

	"finderr/auth"
	"finderr/models"
	"finderr/utils"

	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
	"gorm.io/gorm"
)

var emailPattern = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`,
)

func main() {
	fmt.Println("================================")
	fmt.Println("      MEDIA SERVER SETUP")
	fmt.Println("      Initial Administrator")
	fmt.Println("================================")

	if err := utils.ConnectDB(); err != nil {
		fmt.Fprintln(os.Stderr, "Database initialization failed:", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	name, err := readName(reader)
	if err != nil {
		fail(err)
	}

	email, err := readEmail(reader)
	if err != nil {
		fail(err)
	}

	password, err := readPassword()
	if err != nil {
		fail(err)
	}

	err = validatePassword(password)
	if err != nil {
		fail(err)
	}

	confirmPassword, err := readPasswordConfirmation()
	if err != nil {
		fail(err)
	}

	if password != confirmPassword {
		fail(errors.New("passwords do not match"))
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		fail(fmt.Errorf("password hashing failed: %w", err))
	}

	if err := createSuperAdmin(name, email, passwordHash); err != nil {
		fail(err)
	}

	fmt.Println()
	fmt.Println("================================")
	fmt.Println("Super Admin created successfully!")
	fmt.Println("================================")
	fmt.Println("Name:", name)
	fmt.Println("Email:", email)
	fmt.Println()
	fmt.Println("You can now use these credentials to log in.")
}

func readName(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Administrator name: ")

		name, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		name = strings.TrimSpace(name)

		if name == "" || utf8.RuneCountInString(name) > 150 {
			fmt.Println("Enter a name between 1 and 150 characters.")
			continue
		}

		return name, nil
	}
}

func readEmail(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Administrator email: ")

		email, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		email = strings.ToLower(strings.TrimSpace(email))

		if len(email) > 255 || !emailPattern.MatchString(email) {
			fmt.Println("Please enter a valid email address")
			continue
		}

		return email, nil
	}
}

func readPassword() (string, error) {
	fmt.Print("Administrator password: ")

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return "", err
	}

	return string(password), nil
}

func readPasswordConfirmation() (string, error) {
	fmt.Print("Confirm password: ")

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return "", err
	}

	return string(password), nil
}

func validatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must contain at least 12 bytes")
	}

	if len(password) > 128 {
		return errors.New("password must not exceed 128 bytes")
	}

	return nil
}

func createSuperAdmin(
	name string,
	email string,
	passwordHash string,
) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		//
		//
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(?)",
			int64(824731920),
		).Error; err != nil {
			return fmt.Errorf("acquire setup lock: %w", err)
		}

		var existing int64

		// Check whether a Auper Admin already exists
		if err := tx.Model(&models.User{}).
			Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.name = ?", "super_admin").
			Count(&existing).Error; err != nil {
			return fmt.Errorf("check existing administrators: %w", err)
		}

		if existing > 0 {
			return errors.New(
				"a Super Admin already exists; initial setup is disabled",
			)
		}

		// Check email uniqueness
		var emailCount int64

		if err := tx.Model(&models.User{}).
			Where("LOWER(email) = ?", email).
			Count(&emailCount).Error; err != nil {
			return fmt.Errorf("check email: %w", err)
		}

		if emailCount > 0 {
			return errors.New("an account with this email already exists")
		}

		// Retrieve built-in Super Admin role
		var role models.Role

		if err := tx.Where(
			"name = ? AND is_system = ?",
			"super_admin",
			true,
		).First(&role).Error; err != nil {
			return fmt.Errorf("Super Admin role not found: %w", err)
		}

		// Create administrator account
		user := models.User{
			Username:     name,
			Email:        email,
			PasswordHash: passwordHash,
			IsActive:     true,
			IsVerified:   true,
		}

		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("create administrator: %w", err)
		}

		// Assign
		if err := tx.Model(&user).
			Association("Roles").
			Append(&role); err != nil {
			return fmt.Errorf("assign Super Admin role: %w", err)
		}

		return nil
	})
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "\nSetup failed:", err)
	os.Exit(1)
}
