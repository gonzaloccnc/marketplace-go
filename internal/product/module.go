package product

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires the product feature (repository -> service -> handler)
// and mounts its HTTP routes on the given router.
func Register(r gin.IRouter, pool *pgxpool.Pool) {
	repository := NewProductRepositoryPostgres(pool)
	service := NewProductService(repository)
	handler := NewHTTPProductHandler(service)

	productsGroup := r.Group("/products")
	productsGroup.GET("", handler.GetProducts)
	productsGroup.GET("/:id", handler.GetProductById)
	productsGroup.POST("", handler.CreateProduct)
}
