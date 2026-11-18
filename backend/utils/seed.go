package utils

import (
	"finderr/authorization"
	"finderr/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleDefinition struct {
	Name        string
	Description string
	Permissions []string
}

func SeedRoles(db *gorm.DB) error {
	roles := []RoleDefinition{
		{
			Name: "super_admin",
			Description: "Full control over the media server",
		},
		{
			Name: "administrator",
			Description: "Manages users, media and server settings",
			Permissions: []string{
				authorization.UsersView,
				authorization.UsersCreate,
				authorization.UsersUpdate,
				authorization.UsersDelete,
				authorization.UsersManage,

				authorization.MediaView,
				authorization.MediaUpload,
				authorization.MediaUpdate,
				authorization.MediaDelete,
				authorization.MediaDownload,
				authorization.MediaStream,

				authorization.LibraryView,
				authorization.LibraryScan,
				authorization.LibraryManage,

				authorization.CategoriesView,
				authorization.CategoriesCreate,
				authorization.CategoriesUpate,
				authorization.CategoriesDelete,

				authorization.SettingsView,
				authorization.SettingsUpdate,

				authorization.RolesView,
				authorization.RolesCreate,
				authorization.RolesUpdate,
				authorization.RolesDelete,
				authorization.RolesAssign,

				authorization.AuditLogsView,
			},
		},
		{
			Name: "media_manager",
			Description: "Manage the media library",
			Permissions: []string{
				authorization.MediaView,
				authorization.MediaUpload,
				authorization.MediaUpdate,
				authorization.MediaDelete,
				authorization.MediaDownload,

				authorization.LibraryView,
				authorization.LibraryScan,
				authorization.LibraryManage,

				authorization.CategoriesView,
				authorization.CategoriesCreate,
				authorization.CategoriesUpate,
				authorization.CategoriesDelete,
			},
		},
		{
			Name: "editor",
			Description: "Edit media metadata and categories",
			Permissions: []string{
				authorization.MediaView,
				authorization.MediaUpdate,
				authorization.MediaStream,

				authorization.CategoriesView,
				authorization.CategoriesCreate,
				authorization.CategoriesUpate,
			},
		},
		{
			Name: "viewer",
			Description: "Browse and stream permitted media",
			Permissions: []string{
				authorization.MediaView,
				authorization.MediaStream,
				authorization.LibraryView,
				authorization.CategoriesView,
			},
		},
		{
			Name: "guest",
			Description: "Access explicitly public media only",
			Permissions: []string{
				authorization.MediaView,
				authorization.MediaStream,
			},
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Update the permission catalogue
		for _, definition := range authorization.AllPermissions {
			permission := models.Permission{
				Name: definition.Name,
				Description: definition.Description,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns(
					[]string{"description"},
				),
			}).Create(&permission).Error; err != nil {
				return fmt.Errorf(
					"seed permission %s: %w",
					definition.Name,
					err,
				)
			}
		}

		// Update built-in roles
		for _, definition := range roles {
			role := models.Role{
				Name: definition.Name,
				Description: definition.Description,
				IsSystem: true,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns(
					[]string{"description", "is_system"},
				),
			}).Create(&role).Error; err != nil {
				return fmt.Errorf(
					"seed role %s: %w",
					definition.Name,
					err,
				)
			}

			if err := tx.Where(
				"name = ?", definition.Name,
			).First(&role).Error; err != nil {
				return err
			}

			// Grant every permission to Super Admin
			permissionNames := definition.Permissions

			if definition.Name == "super_admin" {
				permissionNames = make([]string, 0, len(authorization.AllPermissions))

				for _, permission := range authorization.AllPermissions {
					permissionNames = append(permissionNames, permission.Name)
				}
			}

			var permissions []models.Permission

			if len(permissionNames) > 0 {
				if err := tx.Where(
					"name IN ?", permissionNames,
				).Find(&permissions).Error; err != nil {
					return err
				}
			}

			// Synchronize only built-in roles
			if err := tx.Model(&role).Association("Permissions").Replace(permissions); err != nil {
				return fmt.Errorf(
					"assign permissions to %s: %w",
					definition.Name,
					err,
				)
			}
		}

		return nil
	})
}