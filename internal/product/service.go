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
	}

	return &dtoProduct, nil
}

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, product ProductRequest) (*ProductDTO, error) {
	productModel := ProductModel{
		Name:       product.Name,
		Price:      product.Price,
		Stock:      product.Stock,
		Attributes: product.Attributes,
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
	}

	return &dtoProduct, nil
}
