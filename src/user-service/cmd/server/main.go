package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/quochao170402/ecommerce-aws/user-service/configs"
)

func main() {
	cfg, err := configs.LoadConfig()

	if err != nil {
		fmt.Printf("Error when load configuration because %v", err.Error())
		return
	}

	db := cfg.Database
	configs.InitDatabase(db)

	router := gin.Default()
	configs.SetupRoutes(router, cfg)
}
