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
	rows, err := r.connection.Query(ctx, "SELECT id, name, price, stock, attributes, state FROM products WHERE state = $1", ProductStateActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []ProductModel = make([]ProductModel, 0)

	for rows.Next() {
		var product ProductModel
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Attributes, &product.State); err != nil {
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
	row := r.connection.QueryRow(ctx, "SELECT id, name, price, stock, attributes, state FROM products WHERE id = $1", id)
	var product ProductModel
	if err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Attributes, &product.State); err != nil {
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
		"state":      product.State,
	}

	var productInserted *ProductModel = &ProductModel{}

	err := r.connection.QueryRow(
		ctx,
		"INSERT INTO products (name, price, stock, attributes, state) VALUES (@name, @price, @stock, @attributes, @state) RETURNING id, name, price, stock, attributes, state",
		args,
	).Scan(&productInserted.ID, &productInserted.Name, &productInserted.Price, &productInserted.Stock, &productInserted.Attributes, &productInserted.State)

	if err != nil {
		return nil, err
	}

	return productInserted, nil
}

func (r *ProductRepositoryPostgres) InactiveProduct(ctx context.Context, id uuid.UUID) (*ProductModel, error) {
	args := pgx.NamedArgs{
		"id":    id,
		"state": ProductStateInactive,
	}

	var product ProductModel

	err := r.connection.QueryRow(
		ctx,
		"UPDATE products SET state = @state, updated_at = now() WHERE id = @id RETURNING id, name, price, stock, attributes, state",
		args,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Attributes, &product.State)

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepositoryPostgres) UpdateProduct(ctx context.Context, product *ProductModel) (*ProductModel, error) {
	args := pgx.NamedArgs{
		"id":         product.ID,
		"name":       product.Name,
		"price":      product.Price,
		"stock":      product.Stock,
		"attributes": product.Attributes,
		"state":      product.State,
	}

	var productUpdated ProductModel

	err := r.connection.QueryRow(
		ctx,
		"UPDATE products SET name = @name, price = @price, stock = @stock, attributes = @attributes, state = @state, updated_at = now() WHERE id = @id RETURNING id, name, price, stock, attributes, state",
		args,
	).Scan(&productUpdated.ID, &productUpdated.Name, &productUpdated.Price, &productUpdated.Stock, &productUpdated.Attributes, &productUpdated.State)

	if err != nil {
		return nil, err
	}

	return &productUpdated, nil
}
