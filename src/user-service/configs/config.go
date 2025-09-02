package configs

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AppConfig struct {
	AppEnv  string
	AppPort string
}

type Config struct {
	App      AppConfig
	Database *gorm.DB
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
		return nil, err
	}

	database := SetupDatabase()

	appConfig := AppConfig{
		AppEnv:  os.Getenv("APP_ENV"),
		AppPort: os.Getenv("USER_SERVICE_PORT"),
	}

	return &Config{
		App:      appConfig,
		Database: database,
	}, nil
}

func SetupDatabase() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database: " + err.Error())
	}

	return database
}

func InitDatabase(db *gorm.DB) {

	err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.RefreshToken{})

	if err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}

	SeedDatabase(db)
}

func SeedDatabase(db *gorm.DB) {
	var count int64
	db.Model(&models.Role{}).Count(&count)
	if count > 0 {
		return
	}

	roles := []models.Role{
		{ID: uuid.New(), Name: "CUSTOMER", Description: "Default customer role"},
		{ID: uuid.New(), Name: "EMPLOYEE", Description: "Employee role with limited access"},
		{ID: uuid.New(), Name: "ADMIN", Description: "Administrator role with full access"},
	}

	db.Create(roles)
}
