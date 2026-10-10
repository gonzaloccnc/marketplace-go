package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ProductState string

const (
	ProductStateActive   ProductState = "A"
	ProductStateInactive ProductState = "I"
)

type ProductModel struct {
	ID         uuid.UUID
	Name       string
	Price      float64
	Stock      int
	Attributes map[string]string
	State      ProductState
}

type ProductDTO struct {
	ID         string
	Name       string
	Price      float64
	Stock      int
	Attributes map[string]string
	State      ProductState
}

type ProductRequest struct {
	Name       string            `json:"name" binding:"required,min=1"`
	Price      float64           `json:"price" binding:"required,gt=0,numeric"`
	Stock      int               `json:"stock" binding:"required,gt=0,number"`
	Attributes map[string]string `json:"attributes" binding:"required"`
}

// UpdateProductRequest replaces the editable fields of a product. State is
// optional: when omitted the product keeps its current state.
type UpdateProductRequest struct {
	Name       string            `json:"name" binding:"required,min=1"`
	Price      float64           `json:"price" binding:"required,gt=0,numeric"`
	Stock      int               `json:"stock" binding:"required,gt=0,number"`
	Attributes map[string]string `json:"attributes" binding:"required"`
	State      ProductState      `json:"state" binding:"omitempty,oneof=A I"`
}

func (s ProductState) IsValid() bool {
	return s == ProductStateActive || s == ProductStateInactive
}

type ProductRepository interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductModel, error)
	GetProducts(ctx context.Context) ([]ProductModel, error)
	CreateProduct(ctx context.Context, product *ProductModel) (*ProductModel, error)
	UpdateProduct(ctx context.Context, product *ProductModel) (*ProductModel, error)
	InactiveProduct(ctx context.Context, id uuid.UUID) (*ProductModel, error)
}

type ProductService interface {
	GetProducts(ctx context.Context) ([]ProductDTO, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductDTO, error)
	CreateProduct(ctx context.Context, product ProductRequest) (*ProductDTO, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, product UpdateProductRequest) (*ProductDTO, error)
	InactiveProduct(ctx context.Context, id uuid.UUID) (*ProductDTO, error)
}

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductAlreadyExists     = errors.New("product already exists")
	ErrInvalidProduct           = errors.New("invalid product")
	ErrProductNotAvailable      = errors.New("product not available")
	ErrProductOutOfStock        = errors.New("product out of stock")
	ErrProductNotInStock        = errors.New("product not in stock")
	ErrProductIsAlreadyInactive = errors.New("product is already inactive")
	ErrInvalidProductStock      = errors.New("product stock must be greater than zero")
	ErrInvalidProductState      = errors.New("product state must be A (active) or I (inactive)")
)
