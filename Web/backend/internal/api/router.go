package api

import (
	"github.com/Narotan/Web-SIEM/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	api := r.Group("/api")
	api.Use(middleware.BasicAuth())
	{
		api.GET("/health", HealthHandler)
		api.GET("/events", GetEventsHandler)
		api.GET("/stats", GetStatsHandler)
	}
}
