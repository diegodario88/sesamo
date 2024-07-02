package types

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"golang.org/x/crypto/argon2"
)

const saltLength uint32 = 16
const iterations uint32 = 3
const memory uint32 = 64 * 1024
const parallelism uint8 = 2
const keyLength uint32 = 32

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

func (user *User) HashPassword(password string) (encondedHash string, err error) {
	salt, err := generateRandomBytes(saltLength)

	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
