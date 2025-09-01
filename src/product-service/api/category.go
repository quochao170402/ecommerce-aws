package api

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quochao170402/ecommerce-aws/internal/domain"
	"github.com/quochao170402/ecommerce-aws/internal/repository"
	"github.com/quochao170402/ecommerce-aws/middleware"
)

type CategoryRequest struct {
	Name string `json:"name"`
}

type CategoryHandler struct {
	repo repository.BaseRepository[domain.Category]
}

func NewCategoryHandler(repo repository.BaseRepository[domain.Category]) *CategoryHandler {
	return &CategoryHandler{
		repo: repo,
	}
}

func RegisterCategoryRoutes(rg *gin.RouterGroup, repo repository.BaseRepository[domain.Category]) {
	handler := NewCategoryHandler(repo)

	rg.GET("", handler.GetAll)
	rg.POST("", handler.AddCategory)
	rg.GET("/:id", middleware.UUIDParamMiddleware("id"), handler.GetCategoryById)
	rg.PUT("/:id", middleware.UUIDParamMiddleware("id"), handler.UpdateCategory)
	rg.DELETE("/:id", middleware.UUIDParamMiddleware("id"), handler.DeleteCategory)
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.repo.ScanItems(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Data:    categories,
	})
}

func (h *CategoryHandler) AddCategory(c *gin.Context) {
	var request CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	category := domain.Category{
		Id:   uuid.New().String(),
		Name: request.Name,
	}

	if err := h.repo.Save(c, &category); err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Failed to save category",
		})
		return
	}

	c.JSON(http.StatusCreated, BaseResponse{
		Success: true,
		Message: "Category created successfully",
		Data:    category,
	})
}

func (h *CategoryHandler) GetCategoryById(c *gin.Context) {
	id := c.Param("id")

	category, err := h.repo.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Error retrieving category",
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, BaseResponse{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Data:    category,
	})
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var request CategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	category, err := h.repo.FindByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: fmt.Sprintf("Not found category %v", id),
		})
		return
	}

	opts := repository.UpdateOptions{
		ExpressionAttributes: map[string]any{
			"name": request.Name,
		},
		ReturnValues: types.ReturnValueAllNew,
	}

	updated, err := h.repo.Update(c, category, opts)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Message: "Category updated successfully",
		Data:    updated,
	})
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	category, err := h.repo.FindByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Not found category",
		})
		return
	}

	// Handle check products in this category if needed

	err = h.repo.DeleteByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Message: "Delete category successfully",
	})
}
