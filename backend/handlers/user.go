package handlers

import (
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
	
}