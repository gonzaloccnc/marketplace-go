package product

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HTTPProductHandler struct {
	service ProductService
}

func NewHTTPProductHandler(service ProductService) *HTTPProductHandler {
	return &HTTPProductHandler{service: service}
}

func (h *HTTPProductHandler) GetProducts(c *gin.Context) {
	products, err := h.service.GetProducts(c.Request.Context())
	if err != nil {
		slog.Error("failed to get products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *HTTPProductHandler) GetProductById(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		slog.Error("failed to bind uri", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid id, must be uuid",
			"status":  http.StatusBadRequest,
			"message": http.StatusText(http.StatusBadRequest),
		})
		return
	}

	id, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid id, must be uuid",
			"status":  http.StatusBadRequest,
			"message": http.StatusText(http.StatusBadRequest),
		})
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get product by id", "error", err)
		if errors.Is(err, ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "product not found",
				"status":  http.StatusNotFound,
				"message": http.StatusText(http.StatusNotFound),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal server error",
			"status":  http.StatusInternalServerError,
			"message": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *HTTPProductHandler) CreateProduct(c *gin.Context) {
	var product ProductRequest
	if err := c.ShouldBind(&product); err != nil {
		slog.Error("failed to bind product", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid product",
			"status":  http.StatusBadRequest,
			"message": http.StatusText(http.StatusBadRequest),
		})
		return
	}

	productCreated, err := h.service.CreateProduct(c.Request.Context(), product)
	if err != nil {
		slog.Error("failed to create product", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create product",
			"status":  http.StatusInternalServerError,
			"message": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	c.JSON(http.StatusCreated, productCreated)
}
