package handlers

import (
	"errors"
	"finderr/models"
	"finderr/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// Authentication middleware set context
const UserIDContextKey = "user_id"

func getActorID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(UserIDContextKey)
	if !exists {
		return 0, false
	}

	switch id := value.(type) {
	case uint:
		return id, id > 0
	case uint64:
		return uint(id), id > 0
	case int:
		return uint(id), id > 0
	case int64:
		return uint(id), id > 0
	default:
		return 0, false
	}
}

// POST
func (h *Handler) CreateUser(c *gin.Context) {
	actorID, ok := getActorID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		Username string `json:"username" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		RoleNames []string `json:"roles" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := services.CreateInvitation(
		h.DB,
		actorID,
		services.CreateInvitatiionInput{
			Username: input.Username,
			Email: input.Email,
			RoleNames: input.RoleNames,
		},
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created and invitation generated",
		"user_id": result.UserID,
		"email": result.Email,
		"token": result.Token,
		"expires_at": result.ExpiresAt,
	})
}

// GET api
func (h *Handler) ListUsers(c *gin.Context) {
	var users []models.User

	query := h.DB.Preload("Roles").Order("created_at DESC")

	// Optional search by
	search := strings.TrimSpace(c.Query("search"))
	if search != "" {
		query = query.Where(
			"name ILIKE ? OR email ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
		return
	}

	offset := (page - 1) * limit

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	response := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, models.ToUserResponse(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// TO DO: API TO GET USERS
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var user models.User

	if err := h.DB.Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, models.ToUserResponse(user))
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var input struct {
		Username   *string  `json:"username"`
		Email      *string  `json:"email"`
		IsActive   *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := make(map[string]interface{})

	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		if username == "" || len([]rune(username)) > 150 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
			return
		}
		updates["username"] = username
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if email == "" || len(email) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
			return
		}
		updates["email"] = email
	}

	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no changes supplied"})
		return
	}

	result := h.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	actorID, ok := getActorID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if uint(id) == actorID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot delete your own account"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}