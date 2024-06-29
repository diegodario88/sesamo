package store

import (
	"fmt"

	"github.com/diegodario88/sesamo/types"
	"github.com/jmoiron/sqlx"
)

type IUserStore interface {
	FindUserByEmail(email string) (*types.User, error)
}

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) *UserStore {
	var newUserStore = UserStore{
		db: db,
	}

	return &newUserStore
}

func (userStore *UserStore) InsertUser(user *types.User) (*types.User, error) {
	var insertResult types.User
	sqlQuery := `INSERT INTO users (email, password_hash) values ($1, $2) returning *`

	err := userStore.db.Get(&insertResult, sqlQuery, user.Email, user.PasswordHash)

	if err != nil {
		return nil, fmt.Errorf("InsertUser: %w", err)
	}

	return &insertResult, nil
}

func (userStore *UserStore) CountUsers() (int, error) {
	var countResult int
	sqlQuery := `SELECT COUNT(*) FROM users`
	err := userStore.db.QueryRow(sqlQuery).Scan(&countResult)

	return countResult, err
}

func (userStore *UserStore) FindUserByEmail(email string) (*types.User, error) {
	var foundResult types.User
	sqlQuery := `SELECT * FROM users u WHERE u.email = $1`

	err := userStore.db.Get(&foundResult, sqlQuery, email)

	if err != nil {
		return nil, fmt.Errorf("FindUserByEmail: %w", err)
	}

	return &foundResult, nil
}
