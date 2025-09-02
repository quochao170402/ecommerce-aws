package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "supersecret"))

var accessExpire = func() time.Duration {
	minutes, err := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRE", "20"))
	if err != nil {
		minutes = 20
	}
	return time.Duration(minutes) * time.Minute
}()

var refreshExpire = func() time.Duration {
	days, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE", "7"))
	if err != nil {
		days = 7
	}
	return time.Duration(days) * 24 * time.Hour
}()

// Prepare claim names
const (
	ClaimUserID    = "user_id"
	ClaimUserName  = "user_name"
	ClaimUserEmail = "user_email"
	ClaimRole      = "role"
	ClaimExp       = "exp"
)

// Generate access & refresh tokens
func GenerateTokens(user models.User, role string) (string, string, error) {

	accessClaims := jwt.MapClaims{
		ClaimUserID:    user.ID,
		ClaimUserEmail: user.Email,
		ClaimUserName:  user.Name,
		ClaimRole:      user.Role.Name,
		"exp":          time.Now().Add(time.Minute * accessExpire).Unix(),
		"iat":          time.Now().Unix(),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		ClaimUserID: user.ID,
		"exp":       time.Now().Add(refreshExpire * 24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Parse token → return claims
func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}
	return claims, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
