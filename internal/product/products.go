package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ProductModel struct {
	ID         uuid.UUID
	Name       string
	Price      float64
	Stock      int
	Attributes map[string]string
}

type ProductDTO struct {
	ID         string
	Name       string
	Price      float64
	Stock      int
	Attributes map[string]string
}

type ProductRequest struct {
	Name       string            `json:"name" binding:"required,min=1"`
	Price      float64           `json:"price" binding:"required,gt=0,numeric"`
	Stock      int               `json:"stock" binding:"required,gt=0,number"`
	Attributes map[string]string `json:"attributes" binding:"required"`
}

type ProductRepository interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductModel, error)
	GetProducts(ctx context.Context) ([]ProductModel, error)
	CreateProduct(ctx context.Context, product *ProductModel) (*ProductModel, error)
}

type ProductService interface {
	GetProducts(ctx context.Context) ([]ProductDTO, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductDTO, error)
	CreateProduct(ctx context.Context, product ProductRequest) (*ProductDTO, error)
}

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidProduct       = errors.New("invalid product")
	ErrProductNotAvailable  = errors.New("product not available")
	ErrProductOutOfStock    = errors.New("product out of stock")
	ErrProductNotInStock    = errors.New("product not in stock")
)
