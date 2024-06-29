package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
)

type IUser interface {
	CheckPassword(password string) (bool, error)
}

type User struct {
	ID           int       `db:"id"`
	Email        string    `db:"email"`
	PasswordHash *string   `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

var ErrNoPasswordSet = errors.New("no password set for user")
var ErrInvalidUserOrPassword = errors.New("invalid user or password")

func (user *User) CheckPassword(password string) (bool, error) {
	if user.PasswordHash == nil || !strings.HasPrefix(*user.PasswordHash, "$argon2id$") {
		return false, ErrNoPasswordSet
	}

	match, err := argon2id.ComparePasswordAndHash(password, *user.PasswordHash)
	if err != nil {
		return false, fmt.Errorf("CheckPassword: %w", err)
	}
	if !match {
		return false, ErrInvalidUserOrPassword
	}

	return true, nil
}
