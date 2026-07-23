package users

import (
	"context"
	"errors"

	"github.com/gonzaloccnc/marketplace-go/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// credentialsFinder adapts the users repository to security.UserFinder, letting
// the security package authenticate users without depending on this domain.
type credentialsFinder struct {
	repository UserRepository
}

// NewCredentialsFinder returns a security.UserFinder backed by the users store.
func NewCredentialsFinder(pool *pgxpool.Pool) security.UserFinder {
	return &credentialsFinder{repository: NewUserRepository(pool)}
}

// FindByEmail implements [security.UserFinder]. It returns (nil, nil) when no
// user matches, so the auth use case can map it to invalid credentials.
func (f *credentialsFinder) FindByEmail(ctx context.Context, email string) (*security.Credentials, error) {
	user, err := f.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &security.Credentials{
		ID:           user.ID.String(),
		Email:        user.Email,
		Name:         user.Name,
		PasswordHash: user.Password,
	}, nil
}
