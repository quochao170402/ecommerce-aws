package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/repository"
	"github.com/quochao170402/ecommerce-aws/user-service/middleware"
)

type RoleHandler struct {
	repo     repository.IBaseRepository[models.Role]
	userRepo repository.IBaseRepository[models.User]
}

func NewRoleHandler(roleRepo repository.IRoleRepository,
	userRepo repository.IUserRepository) *RoleHandler {
	return &RoleHandler{
		repo:     roleRepo,
		userRepo: userRepo,
	}
}

func RegisterRoleRoutes(rg *gin.RouterGroup, roleRepo repository.IRoleRepository,
	userRepo repository.IUserRepository) {
	RoleHandler := NewRoleHandler(roleRepo, userRepo)
	// rg.Use(middleware.AuthMiddleware(), middleware.RequireRole("ADMIN"))

	rg.GET("", RoleHandler.GetRoles)
	rg.POST("", RoleHandler.AddRole)
	rg.GET("/:id", middleware.UUIDParamMiddleware("id"), RoleHandler.GetRoleById)
	rg.PUT("/:id", middleware.UUIDParamMiddleware("id"), RoleHandler.UpdateRole)
	rg.DELETE("/:id", middleware.UUIDParamMiddleware("id"), RoleHandler.DeleteRole)
}

// GET /roles
func (h *RoleHandler) GetRoles(c *gin.Context) {
	roles, err := h.repo.GetMany(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// POST /roles
func (h *RoleHandler) AddRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// GET /roles/:id
func (h *RoleHandler) GetRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	role, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// PUT /roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role.ID = id // ensure update matches path param
	if err := h.repo.Update(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// DELETE /roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
