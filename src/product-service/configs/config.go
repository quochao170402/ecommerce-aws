package configs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppEnv  string
	AppPort string
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

type Config struct {
	App AppConfig
	AWS aws.Config
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
		return nil, err
	}

	appConfig := AppConfig{
		AppEnv:  os.Getenv("APP_ENV"),
		AppPort: os.Getenv("APP_PORT"),
	}

	// Load the default AWS configuration, which now includes values from .env.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	return &Config{
		App: appConfig,
		AWS: cfg,
	}, nil
}

func LoadDynamoDBConfig() {

}
