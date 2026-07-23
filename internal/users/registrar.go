package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/gonzaloccnc/marketplace-go/internal/security"
	"github.com/jackc/pgx/v5/pgxpool"
)

// registrarAdapter adapts the users service to security.UserRegistrar, so the
// security package can expose /auth/register without depending on this domain.
type registrarAdapter struct {
	service UserService
}

// NewRegistrarAdapter returns a security.UserRegistrar backed by the users store.
func NewRegistrarAdapter(pool *pgxpool.Pool) security.UserRegistrar {
	return &registrarAdapter{service: NewUserServiceImpl(NewUserRepository(pool))}
}

// Register implements [security.UserRegistrar] by delegating to the users
// creation use case and mapping between the boundary DTOs.
func (a *registrarAdapter) Register(ctx context.Context, req security.RegisterRequest) (*security.RegisteredUser, error) {
	created, err := a.service.CreateUser(ctx, &UserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}, CreationOrigin{Via: CreationSelfRegister})
	if err != nil {
		// Translate users domain errors into the security boundary's errors so
		// the auth handler can respond without importing this package.
		switch {
		case errors.Is(err, EmailAlreadyExistsError):
			return nil, security.ErrEmailAlreadyExists
		case errors.Is(err, NameTooShortError), errors.Is(err, PasswordInsufficientError):
			return nil, fmt.Errorf("%w: %s", security.ErrInvalidRegistration, err)
		default:
			return nil, err
		}
	}

	return &security.RegisteredUser{
		ID:    created.ID,
		Name:  created.Name,
		Email: created.Email,
	}, nil
}
