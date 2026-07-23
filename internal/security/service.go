package security

import (
	"context"

	"github.com/gonzaloccnc/marketplace-go/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	finder UserFinder
}

var _ AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(finder UserFinder) AuthService {
	return &AuthServiceImpl{finder: finder}
}

// Authenticate implements [AuthService]. It verifies the credentials and, on
// success, issues a signed JWT. A missing user and a wrong password both
// collapse into InvalidCredentialsError to avoid leaking which emails exist.
func (s *AuthServiceImpl) Authenticate(ctx context.Context, email string, password string) (*AuthResponse, error) {
	creds, err := s.finder.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if creds == nil {
		return nil, InvalidCredentialsError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		return nil, InvalidCredentialsError
	}

	tok, err := token.Generate(creds.ID, creds.Email, creds.Name)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: tok.Value, ExpiresAt: tok.ExpiresAt}, nil
}
