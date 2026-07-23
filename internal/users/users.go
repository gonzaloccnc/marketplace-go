package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// CreationSource records which path created a user. It maps to the
// user_created_via enum in the database.
type CreationSource string

const (
	// CreationSelfRegister marks a user created through the public
	// POST /auth/register endpoint (no acting user).
	CreationSelfRegister CreationSource = "self_register"
	// CreationAdmin marks a user created by an authenticated caller through
	// POST /users; the caller is recorded in CreatedBy.
	CreationAdmin CreationSource = "admin"
)

// CreationOrigin carries the audit context for a user-creation call: which path
// created the user and, when applicable, the authenticated actor who did it.
type CreationOrigin struct {
	Via       CreationSource
	CreatedBy *uuid.UUID // nil for public self-registration
}

type UserModel struct {
	ID         uuid.UUID
	Name       string
	Email      string
	Password   string
	CreatedVia CreationSource
	CreatedBy  *uuid.UUID
	UpdatedBy  *uuid.UUID
}

type UserDTO struct {
	ID         string
	Name       string
	Email      string
	CreatedVia string
	CreatedBy  *string
	UpdatedBy  *string
}

type UserRequest struct {
	Name     string `json:"name" binding:"min=3"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=6"`
}

// AdminUserRequest is the payload for admin-driven creation (POST /users). The
// caller supplies only name and email; the server generates the password.
type AdminUserRequest struct {
	Name  string `json:"name" binding:"min=3"`
	Email string `json:"email" binding:"email"`
}

// AdminCreatedUser is returned when an admin creates a user.
type AdminCreatedUser struct {
	User *UserDTO
}

var (
	// CREATION ERRORS
	NameTooShortError         = errors.New("name must be at least 3 characters")
	EmailAlreadyExistsError   = errors.New("email already exists")
	PasswordInsufficientError = errors.New("password must include at least 6 characters, uppercase, lowercase, and a number")

	// GET ERRORS
	UserNotFoundError = errors.New("user not found")

	// OPERATIONAL ERRORS
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModel, error)
	GetUserByEmail(ctx context.Context, email string) (*UserModel, error)
	CreateUser(ctx context.Context, user *UserModel) (*UserModel, error)
	UpdateUser(ctx context.Context, user *UserModel) (*UserModel, error)
	// DeleteUser soft-deletes the user, recording deletedBy as the acting actor.
	DeleteUser(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
}

type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserDTO, error)
	GetUserByEmail(ctx context.Context, email string) (*UserDTO, error)
	CreateUser(ctx context.Context, user *UserRequest, origin CreationOrigin) (*UserDTO, error)
	CreateUserByAdmin(ctx context.Context, req *AdminUserRequest, createdBy uuid.UUID) (*AdminCreatedUser, error)
	// UpdateUser updates the user, recording updatedBy as the acting actor.
	UpdateUser(ctx context.Context, id uuid.UUID, user *UserRequest, updatedBy uuid.UUID) (*UserDTO, error)
	// DeleteUser soft-deletes the user, recording deletedBy as the acting actor.
	DeleteUser(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
}
