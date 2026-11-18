package authorization

const (
	// User management
	UsersView = "users.view"
	UsersCreate = "users.create"
	UsersUpdate = "users.update"
	UsersDelete = "users.delete"
	UsersManage = "users.manage"

	// Media access
	MediaView = "media.view"
	MediaUpload = "media.upload"
	MediaUpdate = "media.update"
	MediaDelete = "media.delete"
	MediaDownload = "media.download"
	MediaStream = "media.stream"

	// Library management
	LibraryView = "library.view"
	LibraryScan = "library.scan"
	LibraryManage = "library.manage"

	// Categories and metadata
	CategoriesView = "categories.view"
	CategoriesCreate = "categories.create"
	CategoriesUpate = "categories.update"
	CategoriesDelete = "categories.delete"

	// Server settings
	SettingsView = "settings.view"
	SettingsUpdate = "settings.update"

	// Role and permission management
	RolesView = "roles.view"
	RolesCreate = "roles.create"
	RolesUpdate = "roles.update"
	RolesDelete = "roles.delete"
	RolesAssign = "roles.assign"

	// Audit logs
	AuditLogsView = "audit_logs.view"
)

type PermissionDefinition struct {
	Name        string
	Description string
}

var AllPermissions = []PermissionDefinition{
	{UsersView, "View user accounts"},
	{UsersCreate, "Create user accounts"},
	{UsersUpdate, "Update user accounts"},
	{UsersDelete, "Delete user accounts"},
	{UsersManage, "Manage all user accounts"},

	{MediaView, "View media information"},
	{MediaUpload, "Upload media"},
	{MediaUpdate, "Edit media information"},
	{MediaDelete, "Delete media"},
	{MediaDownload, "Download media"},
	{MediaStream, "Stream media"},

	{LibraryView, "View the media library"},
	{LibraryScan, "Scan media directories"},
	{LibraryManage, "Manage media library sources"},

	{CategoriesView, "View categories"},
	{CategoriesCreate, "Create categories"},
	{CategoriesUpate, "Edit categories"},
	{CategoriesDelete, "Delete categories"},

	{SettingsView, "View server settings"},
	{SettingsUpdate, "Change server settings"},

	{RolesView, "View roles"},
	{RolesCreate, "Create custom roles"},
	{RolesUpdate, "Update custom roles"},
	{RolesDelete, "Delete custom roles"},
	{RolesAssign, "Assign roles to users"},

	{AuditLogsView, "View security audit logs"},
}