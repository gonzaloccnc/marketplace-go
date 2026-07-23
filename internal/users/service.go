package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserServiceImpl struct {
	repository UserRepository
}

var _ UserService = (*UserServiceImpl)(nil)

func NewUserServiceImpl(repository UserRepository) UserService {
	return &UserServiceImpl{
		repository: repository,
	}
}

// CreateUser implements [UserService].
func (u *UserServiceImpl) CreateUser(ctx context.Context, user *UserRequest, origin CreationOrigin) (*UserDTO, error) {
	if err := validateUserRequest(user); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	userModel := &UserModel{
		Name:       user.Name,
		Email:      user.Email,
		Password:   hashedPassword,
		CreatedVia: origin.Via,
		CreatedBy:  origin.CreatedBy,
	}

	userCreated, err := u.repository.CreateUser(ctx, userModel)
	if err != nil {
		return nil, err
	}

	return toUserDTO(userCreated), nil
}

// CreateUserByAdmin implements [UserService]. It generates a secure password
// for the new user (the admin never supplies one) and records the acting admin
// as the creator, returning the generated password once so it can be handed off.
func (u *UserServiceImpl) CreateUserByAdmin(ctx context.Context, req *AdminUserRequest, createdBy uuid.UUID) (*AdminCreatedUser, error) {
	password, err := generateSecurePassword()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	dto, err := u.CreateUser(ctx, &UserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}, CreationOrigin{
		Via:       CreationAdmin,
		CreatedBy: &createdBy,
	})
	if err != nil {
		return nil, err
	}

	return &AdminCreatedUser{User: dto}, nil
}

// UpdateUser implements [UserService]. updatedBy is the authenticated actor
// performing the update; it is recorded on the row for audit.
func (u *UserServiceImpl) UpdateUser(ctx context.Context, id uuid.UUID, user *UserRequest, updatedBy uuid.UUID) (*UserDTO, error) {
	if err := validateUserRequest(user); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	userModel := &UserModel{
		ID:        id,
		Name:      user.Name,
		Email:     user.Email,
		Password:  hashedPassword,
		UpdatedBy: &updatedBy,
	}

	userUpdated, err := u.repository.UpdateUser(ctx, userModel)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, UserNotFoundError
		}
		return nil, err
	}

	return toUserDTO(userUpdated), nil
}

// DeleteUser implements [UserService]. The user is soft-deleted, and deletedBy
// is the authenticated actor performing the deletion, recorded for audit.
func (u *UserServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	err := u.repository.DeleteUser(ctx, id, deletedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserNotFoundError
		}
		return err
	}

	return nil
}

// GetUserByEmail implements [UserService].
func (u *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*UserDTO, error) {
	user, err := u.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, UserNotFoundError
		}
		return nil, err
	}

	return toUserDTO(user), nil
}

// GetUserByID implements [UserService].
func (u *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	user, err := u.repository.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, UserNotFoundError
		}
		return nil, err
	}

	return toUserDTO(user), nil
}
