package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Narotan/Web-SIEM/internal/api"
	"github.com/Narotan/Web-SIEM/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.GetConfig()
	log.Printf("Starting SIEM-Web Backend for user: %s", cfg.WebUser)

	router := gin.Default()

	// Setup API routes with authentication
	api.SetupRouter(router)

	// Serve frontend static files (relative to Web/backend)
	frontendDir := "../frontend"
	router.Static("/css", frontendDir+"/css")
	router.Static("/js", frontendDir+"/js")
	router.StaticFile("/", frontendDir+"/index.html")
	router.StaticFile("/index.html", frontendDir+"/index.html")
	router.StaticFile("/login.html", frontendDir+"/login.html")
	router.StaticFile("/events.html", frontendDir+"/events.html")

	// Redirect all other routes to index.html (SPA fallback)
	router.NoRoute(func(c *gin.Context) {
		c.File(frontendDir + "/index.html")
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Web-server is running on http://localhost%s", addr)
	log.Printf("Frontend served from ./Web/frontend/")
	log.Printf("Login credentials: %s / %s", cfg.WebUser, cfg.WebPass)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
