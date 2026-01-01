package api

import (
	"github.com/Narotan/Web-SIEM/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// Apply CORS middleware
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		// Public endpoints (for auth testing)
		api.GET("/health", middleware.BasicAuth(), HealthHandler)

		// Protected endpoints
		protected := api.Group("")
		protected.Use(middleware.BasicAuth())
		{
			protected.GET("/events", GetEventsHandler)
			protected.GET("/stats", GetStatsHandler)
		}
	}
}
