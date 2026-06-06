package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepositoryPostgres struct {
	connection *pgxpool.Pool
}

var _ ProductRepository = (*ProductRepositoryPostgres)(nil)

func NewProductRepositoryPostgres(connection *pgxpool.Pool) ProductRepository {
	return &ProductRepositoryPostgres{connection: connection}
}

func (r *ProductRepositoryPostgres) GetProducts(ctx context.Context) ([]ProductModel, error) {
	rows, err := r.connection.Query(ctx, "SELECT id, name, price, stock, attributes FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []ProductModel = make([]ProductModel, 0)

	for rows.Next() {
		var product ProductModel
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Attributes); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepositoryPostgres) GetProduct(ctx context.Context, id uuid.UUID) (*ProductModel, error) {
	row := r.connection.QueryRow(ctx, "SELECT id, name, price, stock, attributes FROM products WHERE id = $1", id)
	var product ProductModel
	if err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Attributes); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepositoryPostgres) CreateProduct(ctx context.Context, product *ProductModel) (*ProductModel, error) {
	args := pgx.NamedArgs{
		"name":       product.Name,
		"price":      product.Price,
		"stock":      product.Stock,
		"attributes": product.Attributes,
	}

	var productInserted *ProductModel = &ProductModel{}

	err := r.connection.QueryRow(
		ctx,
		"INSERT INTO products (name, price, stock, attributes) VALUES (@name, @price, @stock, @attributes) RETURNING id, name, price, stock, attributes",
		args,
	).Scan(&productInserted.ID, &productInserted.Name, &productInserted.Price, &productInserted.Stock, &productInserted.Attributes)

	if err != nil {
		return nil, err
	}

	return productInserted, nil
}
