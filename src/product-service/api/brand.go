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

type BrandRequest struct {
	Name string `json:"name"`
}

type BrandHandler struct {
	repo repository.BaseRepository[domain.Brand]
}

func NewBrandHandler(repo repository.BaseRepository[domain.Brand]) *BrandHandler {
	return &BrandHandler{
		repo: repo,
	}
}

func RegisterBrandRoutes(rg *gin.RouterGroup, repo repository.BaseRepository[domain.Brand]) {
	handler := NewBrandHandler(repo)

	rg.GET("", handler.GetAll)
	rg.POST("", handler.AddBrand)
	rg.GET("/:id", middleware.UUIDParamMiddleware("id"), handler.GetBrandById)
	rg.PUT("/:id", middleware.UUIDParamMiddleware("id"), handler.UpdateBrand)
	rg.DELETE("/:id", middleware.UUIDParamMiddleware("id"), handler.DeleteBrand)

}

func (h *BrandHandler) GetAll(c *gin.Context) {
	brands, err := h.repo.ScanItems(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Data:    brands,
	})
}

func (h *BrandHandler) AddBrand(c *gin.Context) {
	var request BrandRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	brand := domain.Brand{
		Id:   uuid.New().String(),
		Name: request.Name,
	}

	if err := h.repo.Save(c, &brand); err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Failed to save brand",
		})
		return
	}

	c.JSON(http.StatusCreated, BaseResponse{
		Success: true,
		Message: "Brand created successfully",
		Data:    brand,
	})
}

func (h *BrandHandler) GetBrandById(c *gin.Context) {
	id := c.Param("id")

	brand, err := h.repo.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Error retrieving brand",
		})
		return
	}

	if brand == nil {
		c.JSON(http.StatusNotFound, BaseResponse{
			Success: false,
			Message: "Brand not found",
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Data:    brand,
	})
}

func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	id := c.Param("id")
	var request BrandRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	brand, err := h.repo.FindByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if brand == nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: fmt.Sprintf("Not found brand %v", id),
		})
		return
	}

	opts := repository.UpdateOptions{
		ExpressionAttributes: map[string]any{
			"name": request.Name,
		},
		ReturnValues: types.ReturnValueAllNew,
	}

	updated, err := h.repo.Update(c, brand, opts)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Message: "Brand updated successfully",
		Data:    updated,
	})
}

func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	id := c.Param("id")

	brand, err := h.repo.FindByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if brand == nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{
			Success: false,
			Message: "Not found brand",
		})
		return
	}

	// Handle check products

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
		Message: "Delete brand successfull",
	})
}
