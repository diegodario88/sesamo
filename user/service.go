package user

import (
	"fmt"
	"net/http"

	"github.com/diegodario88/sesamo/httphelper"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type IUserRepository interface {
	FindUserByEmail(email string) (*UserEntity, error)
	InsertUser(user *UserEntity) (*UserEntity, error)
	CountUsers() (int, error)
}

type UserService struct {
	Repository IUserRepository
}

func NewUserService(db *sqlx.DB) UserService {
	var newUserService = UserService{
		Repository: NewUserRepository(db),
	}

	return newUserService
}

func (userService *UserService) Login(w http.ResponseWriter, r *http.Request) {
	var loginUserPayload LoginUserPayload
	if err := httphelper.ParseJSON(r, &loginUserPayload); err != nil {
		httphelper.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := httphelper.Validate.Struct(loginUserPayload); err != nil {
		errors := err.(validator.ValidationErrors)
		httphelper.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	user, err := userService.authenticateUserByEmailPassword(loginUserPayload)

	if err != nil {
		httphelper.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("not found, invalid email or password"),
		)
		return
	}

	token, err := user.CreateJWT()

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (userService *UserService) Register(w http.ResponseWriter, r *http.Request) {
	var registerUserPayload RegisterUserPayload
	if err := httphelper.ParseJSON(r, &registerUserPayload); err != nil {
		httphelper.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := httphelper.Validate.Struct(registerUserPayload); err != nil {
		errors := err.(validator.ValidationErrors)
		httphelper.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	user, err := userService.Repository.FindUserByEmail(registerUserPayload.Email)

	if err == nil {
		httphelper.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("user with email %s already exists", registerUserPayload.Email),
		)
		return
	}

	hashedPassword, err := user.HashPassword(registerUserPayload.Password)

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	userToBeInserted := UserEntity{
		FirstName:    registerUserPayload.FirstName,
		LastName:     registerUserPayload.LastName,
		Email:        registerUserPayload.Email,
		PasswordHash: &hashedPassword,
	}

	insertedUser, err := userService.Repository.InsertUser(&userToBeInserted)

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusCreated, insertedUser)
}

func (userService *UserService) authenticateUserByEmailPassword(
	loginUserPayload LoginUserPayload,
) (*UserEntity, error) {
	user, err := userService.Repository.FindUserByEmail(loginUserPayload.Email)

	if err != nil {
		return nil, err
	}

	if _, err := user.CheckPassword(loginUserPayload.Password); err != nil {
		return nil, err
	}

	return user, nil
}
