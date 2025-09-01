package configs

import (
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/quochao170402/ecommerce-aws/product-service/api"
	"github.com/quochao170402/ecommerce-aws/product-service/internal/domain"
	"github.com/quochao170402/ecommerce-aws/product-service/internal/repository"
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

	client := dynamodb.NewFromConfig(cfg.AWS)

	brandRepo := repository.NewBaseRepository[domain.Brand](client, "Brands")
	categoryRepo := repository.NewBaseRepository[domain.Category](client, "Categories")
	productRepo := repository.NewProductRepository(client)

	v1 := router.Group("/api/v1")
	{
		brands := v1.Group("/brands")
		{
			api.RegisterBrandRoutes(brands, brandRepo)
		}

		categories := v1.Group("/categories")
		{
			api.RegisterCategoryRoutes(categories, categoryRepo)
		}

		products := v1.Group("/products")
		{
			api.RegisterProductRoutes(products, productRepo)
		}
	}

	// // Create repositories
	// taskRepo := repository.NewTaskRepository(db)
	// userRepo := repository.NewBaseRepository[models.User](db)

	// // API v1 group
	// v1 := router.Group("/api/v1")
	// {
	// 	tasks := v1.Group("/tasks")
	// 	{
	// 		api.RegisterTaskRoutes(tasks, taskRepo)
	// 	}

	// 	users := v1.Group("/users")
	// 	{
	// 		api.RegisterUserRoutes(users, userRepo)
	// 	}
	// }
	// Setup task routes

	// api.RegisterUserRoutes(router, userRepo)

	port := cfg.App.AppPort

	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
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
