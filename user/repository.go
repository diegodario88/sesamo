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

func (userRepository *UserRepository) FindUserById(id string) (*UserEntity, error) {
	var foundResult UserEntity
	sqlQuery := `SELECT * FROM users u WHERE u.id = $1`

	err := userRepository.db.Get(&foundResult, sqlQuery, id)

	if err != nil {
		return nil, fmt.Errorf("FindUserById: %w", err)
	}

	return &foundResult, nil
}

func (userRepository *UserRepository) FindAllUsers() ([]UserEntity, error) {
	var foundResult []UserEntity
	sqlQuery := `SELECT * FROM users`

	err := userRepository.db.Select(&foundResult, sqlQuery)

	if err != nil {
		return nil, fmt.Errorf("FindUserById: %w", err)
	}

	return foundResult, nil
}

func (userRepository *UserRepository) GetRoles(userId string) ([]string, error) {
	query := `
		SELECT r.name 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := userRepository.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (userRepository *UserRepository) HasAccess(userID string, permission string) (bool, error) {
	var hasPermission bool
	err := userRepository.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE u.id = $1 AND p.name = $2
		)
	`, userID, permission).Scan(&hasPermission)

	return hasPermission, err
}
