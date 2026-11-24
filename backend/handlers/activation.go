package handlers

import (
	"finderr/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ActiveAccount(c *gin.Context) {
	var input services.ActivateAccountInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := services.ActivateAccount(h.DB, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account activated successfully",
	})
}
