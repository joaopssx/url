package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"urlshortener/config"
	"urlshortener/db"
	"urlshortener/internal/handler"
	"urlshortener/internal/middleware"
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

	urlRepo := repository.NewURLRepository(database)
	urlService := service.NewURLService(urlRepo, cfg.BaseURL)
	urlHandler := handler.NewURLHandler(urlService)

	statsService := service.NewStatsService(urlRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	r := gin.Default()

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitPerMinute)
	r.Use(rateLimiter.Middleware())

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

	r.POST("/shorten", middleware.OptionalAuth(authService), urlHandler.Shorten)
	r.GET("/:code", urlHandler.Redirect)
	r.GET("/:code/stats", statsHandler.GetStats)

	protectedRoutes := r.Group("/")
	protectedRoutes.Use(middleware.AuthRequired(authService))
	{
		protectedRoutes.GET("/me/urls", urlHandler.GetUserURLs)
		protectedRoutes.DELETE("/:code", urlHandler.DeleteURL)
		protectedRoutes.PATCH("/:code", urlHandler.UpdateURL)
	}

	log.Printf("Starting server on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
