package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quochao170402/ecommerce-aws/user-service/auth"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo         repository.IUserRepository
	roleRepo         repository.IRoleRepository
	refreshTokenRepo repository.IRefreshTokenRepository
}

func NewAuthHandler(userRepo repository.IUserRepository,
	roleRepo repository.IRoleRepository,
	refreshTokenRepo repository.IRefreshTokenRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func RegisterAuthRoutes(rg *gin.RouterGroup,
	userRepo repository.IUserRepository,
	roleRepo repository.IRoleRepository,
	refreshTokenRepo repository.IRefreshTokenRepository) {

	handler := NewAuthHandler(userRepo, roleRepo, refreshTokenRepo)

	rg.POST("/login", handler.Login)
	rg.POST("/register", handler.Register)
	rg.POST("/reset-password", handler.ResetPassword)
	rg.POST("/forgot-password", handler.ForgotPassword)
	rg.POST("/refresh-token", handler.RefreshToken)
}

// ---------- LOGIN ----------
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(*user, user.Role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.refreshTokenRepo.Create(c.Request.Context(), &models.RefreshToken{
		Token: refreshToken,
	})

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// ---------- REGISTER ----------
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	customerRole, err := h.roleRepo.GetByName(c.Request.Context(), "CUSTOMER")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get customer role"})
		return
	}

	if customerRole == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to init customer role for user"})
		return
	}

	user := models.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: string(hashed),
		Name:     req.Name,
		RoleID:   customerRole.ID,
	}

	if err := h.userRepo.Create(context.Background(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// ---------- REFRESH TOKEN ----------
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ParseToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, ok := claims[auth.ClaimUserID].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userRepo.GetByID(context.Background(), uid)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(*user, user.Role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// ---------- FORGOT PASSWORD ----------
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Here you would generate a reset token and send email
	// Skipping email service for now
	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent (stub)"})
}

// ---------- RESET PASSWORD ----------
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user.Password = string(hashed)
	user.LatestUpdatedAt = time.Now()

	if err := h.userRepo.Update(context.Background(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
