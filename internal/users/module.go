package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gonzaloccnc/marketplace-go/internal/token"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register wires the users feature (repository -> service -> handler) and mounts
// its HTTP routes. Registration is public; user management sits behind the
// security auth middleware so we know which authenticated actor is acting.
func Register(r gin.IRouter, pool *pgxpool.Pool) {
	repository := NewUserRepository(pool)
	service := NewUserServiceImpl(repository)
	handler := NewHTTPUserHandler(service)

	// User management: requires authentication and audits who is acting.
	// (Public self-service registration lives in the security module at
	// POST /auth/register, delegating back here via the UserRegistrar port.)
	usersGroup := r.Group("/users")
	usersGroup.Use(token.Middleware())
	usersGroup.POST("", handler.CreateUser)
	usersGroup.GET("/:id", handler.GetUserByID)
	usersGroup.PUT("/:id", handler.UpdateUser)
	usersGroup.DELETE("/:id", handler.DeleteUser)
}
