package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"urlshortener/config"
	"urlshortener/db"
	"urlshortener/internal/handler"
	"urlshortener/internal/repository"
	"urlshortener/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	log.Printf("Starting server on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
