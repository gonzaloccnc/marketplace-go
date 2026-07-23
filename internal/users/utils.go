package users

import (
	"crypto/rand"
	"math/big"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the work factor used when hashing passwords. 12 is a sane
// production default; raise it as hardware improves.
const bcryptCost = 12

// generatedPasswordLength is the length of admin-generated temporary passwords.
const generatedPasswordLength = 20

// randChar returns a cryptographically random character from set.
func randChar(set string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
	if err != nil {
		return 0, err
	}
	return set[n.Int64()], nil
}

// generateSecurePassword returns a cryptographically random password that is
// guaranteed to satisfy validateUserRequest (length, upper, lower, digit). It is
// used when an admin creates a user via POST /users, since the caller does not
// supply a password. Ambiguous-looking characters are excluded on purpose.
func generateSecurePassword() (string, error) {
	const (
		lower   = "abcdefghijkmnopqrstuvwxyz"
		upper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		digits  = "23456789"
		symbols = "!@#$%*-_"
	)
	all := lower + upper + digits + symbols

	pw := make([]byte, 0, generatedPasswordLength)
	// Guarantee at least one character from each required class.
	for _, class := range []string{lower, upper, digits} {
		ch, err := randChar(class)
		if err != nil {
			return "", err
		}
		pw = append(pw, ch)
	}
	for len(pw) < generatedPasswordLength {
		ch, err := randChar(all)
		if err != nil {
			return "", err
		}
		pw = append(pw, ch)
	}

	// Shuffle so the guaranteed characters are not always in the same position.
	for i := len(pw) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		jj := int(j.Int64())
		pw[i], pw[jj] = pw[jj], pw[i]
	}

	return string(pw), nil
}

func containsUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, c := range s {
		if unicode.IsNumber(c) {
			return true
		}
	}
	return false
}

// validateUserRequest enforces the business rules shared by user creation and
// updates before the request reaches the database.
func validateUserRequest(user *UserRequest) error {
	if len(user.Name) < 3 {
		return NameTooShortError
	}

	if len(user.Password) < 6 {
		return PasswordInsufficientError
	}

	if !containsUppercase(user.Password) || !containsLowercase(user.Password) || !containsNumber(user.Password) {
		return PasswordInsufficientError
	}

	return nil
}

// hashPassword returns a bcrypt hash of the given plaintext password.
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// toUserDTO maps a persistence model to the public DTO, dropping the password.
func toUserDTO(user *UserModel) *UserDTO {
	dto := &UserDTO{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		CreatedVia: string(user.CreatedVia),
	}
	if user.CreatedBy != nil {
		createdBy := user.CreatedBy.String()
		dto.CreatedBy = &createdBy
	}
	if user.UpdatedBy != nil {
		updatedBy := user.UpdatedBy.String()
		dto.UpdatedBy = &updatedBy
	}
	return dto
}
