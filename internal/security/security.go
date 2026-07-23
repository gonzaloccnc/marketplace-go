package security

import (
	"context"
	"errors"
)

// AuthResponse is returned to a client after a successful authentication.
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// LoginRequest is the credential payload accepted by the login endpoint.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is the payload accepted by the public registration endpoint.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisteredUser is the public view returned after a successful registration.
type RegisteredUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	// InvalidCredentialsError is returned by Authenticate when the email is
	// unknown or the password does not match.
	InvalidCredentialsError = errors.New("invalid email or password")

	// ErrEmailAlreadyExists is returned by a UserRegistrar when the requested
	// email is already taken. The registration handler maps it to 409 Conflict.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidRegistration is returned by a UserRegistrar when the payload
	// violates the users domain rules (e.g. a weak password) beyond what the
	// binding tags catch. The handler maps it to 400 Bad Request.
	ErrInvalidRegistration = errors.New("invalid registration data")
)

// Credentials is the minimal user data the auth use case needs to verify a
// login. A feature package (users) provides an adapter that produces it, so
// security never depends on the users domain.
type Credentials struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
}

// UserFinder is the port the auth use case depends on. Implementations return
// (nil, nil) when no user matches the email.
type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (*Credentials, error)
}

// UserRegistrar is the port the registration endpoint depends on. The users
// domain provides the adapter so security never imports it directly.
type UserRegistrar interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisteredUser, error)
}

// AuthService is the authentication use-case boundary.
type AuthService interface {
	Authenticate(ctx context.Context, email string, password string) (*AuthResponse, error)
}
