package handlers

import (
	"finderr/models"
	"finderr/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-contrib/sessions"
)

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Check if the user exists in the database
	var user models.User

	result := utils.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	session := sessions.Default(c)

	session.Set("user", user.ID)
	session.Set("role", user.Role)

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}
