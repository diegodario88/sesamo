package user

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

type UserEntity struct {
	ID           string    `db:"id"            json:"id"`
	FirstName    string    `db:"first_name"    json:"firstName"`
	LastName     string    `db:"last_name"     json:"lastName"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash *string   `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type OrganizationEntity struct {
	ID                   string    `db:"id"                      json:"id"`
	ExternalCompanyId    int       `db:"external_company_id"     json:"external_company_id"`
	ExternalHeadOfficeId int       `db:"external_head_office_id" json:"external_head_office_id"`
	Name                 string    `db:"name"                    json:"name"`
	Description          string    `db:"description"             json:"description"`
	CreatedAt            time.Time `db:"created_at"              json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"              json:"updated_at"`
}

type BranchEntity struct {
	ID               string    `db:"id"                 json:"id"`
	ExternalOfficeId int       `db:"external_office_id" json:"external_office_id"`
	CNPJ             string    `db:"cnpj"               json:"cnpj"`
	OrganizationId   string    `db:"organization_id"    json:"organization_id"`
	Name             string    `db:"name"               json:"name"`
	Description      string    `db:"description"        json:"description"`
	IsWarehouse      bool      `db:"is_warehouse"       json:"is_warehouse"`
	CreatedAt        time.Time `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updated_at"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName"  validate:"required"`
	Email     string `json:"email"     validate:"required,email"`
	Password  string `json:"password"  validate:"required,min=3,max=130"`
}

type LoginUserPayload struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

var ErrNoPasswordSet = errors.New("no password set for user")
var ErrInvalidUserOrPassword = errors.New("invalid user or password")

func (user *UserEntity) HashPassword(password string) (encondedHash string, err error) {
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

func (user *UserEntity) CheckPassword(password string) (bool, error) {
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

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
