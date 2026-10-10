package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductServiceImpl struct {
	repository ProductRepository
}

var _ ProductService = (*ProductServiceImpl)(nil)

func NewProductService(repository ProductRepository) ProductService {
	return &ProductServiceImpl{repository: repository}
}

func (s *ProductServiceImpl) GetProducts(ctx context.Context) ([]ProductDTO, error) {
	products, err := s.repository.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	dtoProducts := make([]ProductDTO, 0, len(products))
	for _, product := range products {
		dtoProducts = append(dtoProducts, ProductDTO{
			ID:         product.ID.String(),
			Name:       product.Name,
			Price:      product.Price,
			Stock:      product.Stock,
			Attributes: product.Attributes,
			State:      product.State,
		})
	}
	return dtoProducts, nil
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, id uuid.UUID) (*ProductDTO, error) {
	product, err := s.repository.GetProduct(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	dtoProduct := ProductDTO{
		ID:         product.ID.String(),
		Name:       product.Name,
		Price:      product.Price,
		Stock:      product.Stock,
		Attributes: product.Attributes,
		State:      product.State,
	}

	return &dtoProduct, nil
}

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, product ProductRequest) (*ProductDTO, error) {
	productModel := ProductModel{
		Name:       product.Name,
		Price:      product.Price,
		Stock:      product.Stock,
		Attributes: product.Attributes,
		State:      ProductStateActive,
	}

	productCreated, err := s.repository.CreateProduct(ctx, &productModel)
	if err != nil {
		return nil, err
	}

	dtoProduct := ProductDTO{
		ID:         productCreated.ID.String(),
		Name:       productCreated.Name,
		Price:      productCreated.Price,
		Stock:      productCreated.Stock,
		Attributes: productCreated.Attributes,
		State:      productCreated.State,
	}

	return &dtoProduct, nil
}

func (s *ProductServiceImpl) InactiveProduct(ctx context.Context, id uuid.UUID) (*ProductDTO, error) {
	product, err := s.repository.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if product.State == ProductStateInactive {
		return nil, ErrProductIsAlreadyInactive
	}

	productInactivated, err := s.repository.InactiveProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	dtoProduct := ProductDTO{
		ID:         productInactivated.ID.String(),
		Name:       productInactivated.Name,
		Price:      productInactivated.Price,
		Stock:      productInactivated.Stock,
		Attributes: productInactivated.Attributes,
		State:      productInactivated.State,
	}

	return &dtoProduct, nil
}

func (s *ProductServiceImpl) UpdateProduct(ctx context.Context, id uuid.UUID, product UpdateProductRequest) (*ProductDTO, error) {
	if product.Stock <= 0 {
		return nil, ErrInvalidProductStock
	}

	if product.State != "" && !product.State.IsValid() {
		return nil, ErrInvalidProductState
	}

	current, err := s.repository.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	state := current.State
	if product.State != "" {
		state = product.State
	}

	productModel := ProductModel{
		ID:         id,
		Name:       product.Name,
		Price:      product.Price,
		Stock:      product.Stock,
		Attributes: product.Attributes,
		State:      state,
	}

	productUpdated, err := s.repository.UpdateProduct(ctx, &productModel)
	if err != nil {
		// the product can be deleted between the lookup and the update
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	dtoProduct := ProductDTO{
		ID:         productUpdated.ID.String(),
		Name:       productUpdated.Name,
		Price:      productUpdated.Price,
		Stock:      productUpdated.Stock,
		Attributes: productUpdated.Attributes,
		State:      productUpdated.State,
	}

	return &dtoProduct, nil
}
