package service

import (
	"github.com/diegodario88/sesamo/store"
	"github.com/diegodario88/sesamo/types"
	"github.com/jmoiron/sqlx"
)

type UserService struct {
	store store.IUserStore
}

func NewUserService(db *sqlx.DB) UserService {
	var newUserService = UserService{
		store: store.NewUserStore(db),
	}

	return newUserService
}

func (userService *UserService) AuthenticateUserByEmailPassword(
	email string,
	password string,
) (*types.User, error) {
	user, err := userService.store.FindUserByEmail(email)

	if err != nil {
		return nil, err
	}

	if _, err := user.CheckPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}
