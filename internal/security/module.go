package security

import (
	"github.com/gin-gonic/gin"
)

// Register wires the security feature (auth service -> handler) and mounts its
// public HTTP routes. The UserFinder and UserRegistrar ports are injected by the
// composition root, so security stays free of any dependency on feature
// packages like users.
func Register(r gin.IRouter, finder UserFinder, registrar UserRegistrar) {
	service := NewAuthService(finder)
	handler := NewHTTPAuthHandler(service, registrar)

	authGroup := r.Group("/auth")
	authGroup.POST("/login", handler.Login)
	authGroup.POST("/register", handler.Register)
}
