package middleware

import (
	"net/http"

	"github.com/Narotan/Web-SIEM/internal/config"
	"github.com/gin-gonic/gin"
)

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()

		user, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth || user != cfg.WebUser || password != cfg.WebPass {
			// Skip WWW-Authenticate header to avoid browser auth prompt
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status": "error",
			})
			return
		}

		c.Next()
	}
}
