package user

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	var newUserRepository = UserRepository{
		db: db,
	}

	return &newUserRepository
}

func (userRepository *UserRepository) InsertUser(user *UserEntity) (*UserEntity, error) {
	var insertResult UserEntity
	sqlQuery := `INSERT INTO users (first_name, last_name, email, password_hash) 
                          values ($1, $2, $3, $4) returning *`

	err := userRepository.db.Get(
		&insertResult,
		sqlQuery,
		user.FirstName,
		user.LastName,
		user.Email,
		user.PasswordHash,
	)

	if err != nil {
		return nil, fmt.Errorf("InsertUser: %w", err)
	}

	return &insertResult, nil
}

func (userRepository *UserRepository) CountUsers() (int, error) {
	var countResult int
	sqlQuery := `SELECT COUNT(*) FROM users`
	err := userRepository.db.QueryRow(sqlQuery).Scan(&countResult)

	return countResult, err
}

func (userRepository *UserRepository) FindUserByEmail(email string) (*UserEntity, error) {
	var foundResult UserEntity
	sqlQuery := `SELECT * FROM users u WHERE u.email = $1`

	err := userRepository.db.Get(&foundResult, sqlQuery, email)

	if err != nil {
		return nil, fmt.Errorf("FindUserByEmail: %w", err)
	}

	return &foundResult, nil
}
