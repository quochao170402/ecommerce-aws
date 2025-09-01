package main

import (
	"github.com/gin-gonic/gin"
	"github.com/quochao170402/ecommerce-aws/configs"
)

func main() {
	cfg, err := configs.LoadConfig()

	if err != nil {
		panic("Error load .env file")
	}

	router := gin.New()
	router.Use(gin.Recovery())
	configs.SetupRoutes(router, cfg)
}
