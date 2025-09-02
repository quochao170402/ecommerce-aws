package configs

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/handler"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/repository"
)

func SetupRoutes(router *gin.Engine, cfg *Config) {
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})

	roleRepo := repository.NewRoleRepository(cfg.Database)
	userRepo := repository.NewUserRepository(cfg.Database)
	refreshToken := repository.NewRefreshTokenRepository(cfg.Database)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{

			handler.RegisterAuthRoutes(auth, userRepo, roleRepo, refreshToken)
		}

		roles := v1.Group("/roles")
		{
			handler.RegisterRoleRoutes(roles, roleRepo, userRepo)
		}

	}

	port := cfg.App.AppPort

	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s", port)
	fmt.Println(router.Run(":" + port))
}

func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}
