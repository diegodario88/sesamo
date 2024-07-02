package service

import (
	"github.com/diegodario88/sesamo/store"
	"github.com/diegodario88/sesamo/types"
	"github.com/jmoiron/sqlx"
)

type userFindable interface {
	FindUserByEmail(email string) (*types.User, error)
}

type UserService struct {
	Store userFindable
}

func NewUserService(db *sqlx.DB) UserService {
	var newUserService = UserService{
		Store: store.NewUserStore(db),
	}

	return newUserService
}

func (userService *UserService) AuthenticateUserByEmailPassword(
	email string,
	password string,
) (*types.User, error) {
	user, err := userService.Store.FindUserByEmail(email)

	if err != nil {
		return nil, err
	}

	if _, err := user.CheckPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}
