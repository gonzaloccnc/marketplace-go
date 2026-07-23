package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolationCode is the PostgreSQL SQLSTATE for a unique_violation.
const uniqueViolationCode = "23505"

// isUniqueViolation reports whether err is a PostgreSQL unique constraint
// violation, which lets the repository translate races on the email uniqueness
// index into a domain error instead of leaking the driver error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

type UserRepositoryPostgres struct {
	connection *pgxpool.Pool
}

var _ UserRepository = (*UserRepositoryPostgres)(nil)

func NewUserRepository(connection *pgxpool.Pool) UserRepository {
	return &UserRepositoryPostgres{connection: connection}
}

// CreateUser implements [UserRepository].
func (u *UserRepositoryPostgres) CreateUser(ctx context.Context, user *UserModel) (*UserModel, error) {
	userExists, err := u.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if userExists != nil {
		return nil, EmailAlreadyExistsError
	}

	var args = pgx.NamedArgs{
		"name":        user.Name,
		"email":       user.Email,
		"password":    user.Password,
		"created_via": string(user.CreatedVia),
		"created_by":  user.CreatedBy,
	}

	var insertedUser UserModel
	err = u.connection.QueryRow(
		ctx,
		`INSERT INTO users (name, email, password, created_via, created_by)
		 VALUES (@name, @email, @password, @created_via, @created_by)
		 RETURNING id, name, email, password, created_via, created_by`,
		args,
	).Scan(
		&insertedUser.ID,
		&insertedUser.Name,
		&insertedUser.Email,
		&insertedUser.Password,
		&insertedUser.CreatedVia,
		&insertedUser.CreatedBy,
	)

	if err != nil {
		// Guards against the race between the check above and the insert:
		// two concurrent requests can both pass GetUserByEmail.
		if isUniqueViolation(err) {
			return nil, EmailAlreadyExistsError
		}
		return nil, err
	}

	return &insertedUser, nil
}

// DeleteUser implements [UserRepository]. It performs a soft delete: the row is
// kept and stamped with deleted_at / deleted_by rather than removed, so history
// and audit are preserved. Already-deleted or missing users report no rows.
func (u *UserRepositoryPostgres) DeleteUser(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	args := pgx.NamedArgs{
		"id":         id,
		"deleted_by": deletedBy,
	}

	tag, err := u.connection.Exec(
		ctx,
		`UPDATE users
		    SET deleted_at = now(), deleted_by = @deleted_by
		  WHERE id = @id AND deleted_at IS NULL`,
		args,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetUserByEmail implements [UserRepository].
func (u *UserRepositoryPostgres) GetUserByEmail(ctx context.Context, email string) (*UserModel, error) {
	var user UserModel
	err := u.connection.QueryRow(
		ctx,
		"SELECT id, name, email, password, created_via, created_by, updated_by FROM users WHERE email = $1 AND deleted_at IS NULL",
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedVia, &user.CreatedBy, &user.UpdatedBy)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID implements [UserRepository].
func (u *UserRepositoryPostgres) GetUserByID(ctx context.Context, id uuid.UUID) (*UserModel, error) {
	var user UserModel
	err := u.connection.QueryRow(
		ctx,
		"SELECT id, name, email, password, created_via, created_by, updated_by FROM users WHERE id = $1 AND deleted_at IS NULL",
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedVia, &user.CreatedBy, &user.UpdatedBy)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser implements [UserRepository].
func (u *UserRepositoryPostgres) UpdateUser(ctx context.Context, user *UserModel) (*UserModel, error) {
	args := pgx.NamedArgs{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"password":   user.Password,
		"updated_by": user.UpdatedBy,
	}

	var updatedUser UserModel
	err := u.connection.QueryRow(
		ctx,
		`UPDATE users
		    SET name = @name, email = @email, password = @password,
		        updated_by = @updated_by, updated_at = now()
		  WHERE id = @id AND deleted_at IS NULL
		RETURNING id, name, email, password, created_via, created_by, updated_by`,
		args,
	).Scan(&updatedUser.ID, &updatedUser.Name, &updatedUser.Email, &updatedUser.Password, &updatedUser.CreatedVia, &updatedUser.CreatedBy, &updatedUser.UpdatedBy)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, EmailAlreadyExistsError
		}
		return nil, err
	}

	return &updatedUser, nil
}
