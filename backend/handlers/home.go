package handlers

import "github.com/gin-gonic/gin"

func Home(c *gin.Context) {
	c.HTML(200, "home.html", gin.H{
		"Title": "Home",
		// "Style": "home.css",
		// "Script": "home.js",
	})
}
