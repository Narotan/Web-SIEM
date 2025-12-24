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

	// Serve frontend static files
	router.Static("/css", "../frontend/css")
	router.Static("/js", "../frontend/js")
	router.StaticFile("/", "../frontend/index.html")
	router.StaticFile("/index.html", "../frontend/index.html")
	router.StaticFile("/login.html", "../frontend/login.html")
	router.StaticFile("/events.html", "../frontend/events.html")

	// Redirect root to login if not authenticated (handled by frontend JS)
	router.NoRoute(func(c *gin.Context) {
		c.File("../frontend/index.html")
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Web-server is running on http://localhost%s", addr)
	log.Printf("Frontend served from ../frontend/")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
