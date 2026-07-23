package product

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gonzaloccnc/marketplace-go/pkg/httpx"
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
		httpx.WriteError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteSuccess(c, http.StatusOK, products)
}

func (h *HTTPProductHandler) GetProductById(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		slog.Error("failed to bind uri", "error", err.Error())
		httpx.WriteError(c, http.StatusBadRequest, "invalid id, must be uuid")
		return
	}

	id, err := uuid.Parse(uri.ID)
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid id, must be uuid")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get product by id", "error", err)
		if errors.Is(err, ErrProductNotFound) {
			httpx.WriteError(c, http.StatusNotFound, "product not found")
			return
		}

		httpx.WriteError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteSuccess(c, http.StatusOK, product)
}

func (h *HTTPProductHandler) CreateProduct(c *gin.Context) {
	var product ProductRequest
	if err := c.ShouldBind(&product); err != nil {
		slog.Error("failed to bind product", "error", err.Error())
		httpx.WriteBindError(c, err)
		return
	}

	productCreated, err := h.service.CreateProduct(c.Request.Context(), product)
	if err != nil {
		slog.Error("failed to create product", "error", err.Error())
		httpx.WriteError(c, http.StatusInternalServerError, "failed to create product")
		return
	}

	httpx.WriteSuccess(c, http.StatusCreated, productCreated)
}
