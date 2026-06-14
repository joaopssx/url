package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	
	"urlshortener/config"
	"urlshortener/db"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Database
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Gin router
	r := gin.Default()

	// Health check route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Start the server
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
