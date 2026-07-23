package token

import (
	"errors"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	env "github.com/gonzaloccnc/marketplace-go/config"
)

const (
	issuer = "marketplace-go"
	ttl    = 24 * time.Hour
)

var (
	ErrSecretMissing = errors.New("jwt secret is required")
	ErrInvalidToken  = errors.New("invalid or expired token")
)

// Claims are the claims embedded into every access token.
type Claims struct {
	jwt.Claims
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Token is a signed access token together with its expiry (unix seconds).
type Token struct {
	Value     string
	ExpiresAt int64
}

func secret() ([]byte, error) {
	s := env.GetOrDefault("JWT_SECRET", "")
	if s == "" {
		return nil, ErrSecretMissing
	}
	return []byte(s), nil
}

// Generate issues a signed HS256 JWT for the given identity.
func Generate(subject, email, name string) (*Token, error) {
	key, err := secret()
	if err != nil {
		return nil, err
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiry := now.Add(ttl)

	claims := Claims{
		Claims: jwt.Claims{
			Subject:   subject,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Expiry:    jwt.NewNumericDate(expiry),
		},
		Email: email,
		Name:  name,
	}

	raw, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return nil, err
	}

	return &Token{Value: raw, ExpiresAt: expiry.Unix()}, nil
}

// Validate verifies the signature and standard claims of a JWT and returns the
// parsed claims when the token is valid.
func Validate(raw string) (*Claims, error) {
	key, err := secret()
	if err != nil {
		return nil, err
	}

	parsed, err := jwt.ParseSigned(raw)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims := &Claims{}
	if err := parsed.Claims(key, claims); err != nil {
		return nil, ErrInvalidToken
	}

	if err := claims.Validate(jwt.Expected{Issuer: issuer, Time: time.Now()}); err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
