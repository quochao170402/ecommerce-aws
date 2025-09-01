package api

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quochao170402/ecommerce-aws/product-service/internal/domain"
	"github.com/quochao170402/ecommerce-aws/product-service/internal/repository"
	"github.com/quochao170402/ecommerce-aws/product-service/middleware"
)

type ProductRequest struct {
	Name       string  `json:"name"`
	BrandId    string  `json:"brandId"`
	CategoryId string  `json:"categoryId"`
	Price      float64 `json:"price"`
}

type ProductHandler struct {
	repo repository.ProductRepository
}

func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func RegisterProductRoutes(rg *gin.RouterGroup, repo repository.ProductRepository) {
	handler := NewProductHandler(repo)

	rg.GET("", handler.GetAll)
	rg.POST("", handler.AddProduct)
	rg.GET("/:id", middleware.UUIDParamMiddleware("id"), handler.GetProductById)
	rg.PUT("/:id", middleware.UUIDParamMiddleware("id"), handler.UpdateProduct)
	rg.DELETE("/:id", middleware.UUIDParamMiddleware("id"), handler.DeleteProduct)

	// optional: expose your custom repo methods
	rg.GET("/brand/:brandId", handler.GetByBrand)
	rg.GET("/category/:categoryId", handler.GetByCategory)
	rg.GET("/search", handler.SearchByName)
}

// ------------------ Handlers ------------------

func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.repo.ScanItems(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: products})
}

func (h *ProductHandler) AddProduct(c *gin.Context) {
	var request ProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{Success: false, Message: "Invalid request body"})
		return
	}

	product := domain.Product{
		ID:         uuid.New().String(),
		Name:       request.Name,
		BrandID:    request.BrandId,
		CategoryID: request.CategoryId,
		Price:      request.Price,
	}

	if err := h.repo.Save(c, &product); err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: "Failed to save product"})
		return
	}

	c.JSON(http.StatusCreated, BaseResponse{Success: true, Message: "Product created successfully", Data: product})
}

func (h *ProductHandler) GetProductById(c *gin.Context) {
	id := c.Param("id")

	product, err := h.repo.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, BaseResponse{Success: false, Message: "Product not found"})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: product})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var request ProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, BaseResponse{Success: false, Message: "Invalid request body"})
		return
	}

	product, err := h.repo.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, BaseResponse{Success: false, Message: fmt.Sprintf("Not found product %v", id)})
		return
	}

	opts := repository.UpdateOptions{
		ExpressionAttributes: map[string]any{
			"name":       request.Name,
			"brandId":    request.BrandId,
			"categoryId": request.CategoryId,
			"price":      request.Price,
		},
		ReturnValues: types.ReturnValueAllNew,
	}

	updated, err := h.repo.Update(c, product, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{Success: true, Message: "Product updated successfully", Data: updated})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.repo.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, BaseResponse{Success: false, Message: "Product not found"})
		return
	}

	if err := h.repo.DeleteByID(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{Success: true, Message: "Delete product successfully"})
}

// ------------------ Extra queries ------------------
func (h *ProductHandler) GetByBrand(c *gin.Context) {
	brandId := c.Param("brandId")
	products, err := h.repo.FindByBrand(c, brandId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: products})
}

func (h *ProductHandler) GetByCategory(c *gin.Context) {
	categoryId := c.Param("categoryId")
	products, err := h.repo.FindByCategory(c, categoryId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: products})
}

func (h *ProductHandler) SearchByName(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, BaseResponse{Success: false, Message: "Missing search query parameter ?q="})
		return
	}

	fmt.Println("Queryr: ", keyword)

	products, err := h.repo.SearchByName(c, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BaseResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: products})
}
