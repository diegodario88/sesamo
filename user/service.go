package user

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/diegodario88/sesamo/config"
	"github.com/diegodario88/sesamo/httphelper"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type IUserRepository interface {
	FindUserByEmail(email string) (*UserEntity, error)
	InsertUser(user *UserEntity) (*UserEntity, error)
	CountUsers() (int, error)
	FindUserById(id string) (*UserEntity, error)
	FindAllUsers() ([]UserEntity, error)
	HasAccess(userId string, permission string) (bool, error)
	GetRoles(userId string) ([]string, error)
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
	log.Print("Starting login request ...")
	var loginUserPayload LoginUserPayload
	if err := httphelper.ParseJSON(r, &loginUserPayload); err != nil {
		log.Print(err)
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

	token, err := userService.GenerateUserToken(user)

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (userService *UserService) Register(w http.ResponseWriter, r *http.Request) {
	log.Print("Starting register request ...")
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

func (userService *UserService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := userService.Repository.FindAllUsers()
	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, users)
}

func (userService *UserService) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		httphelper.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	user, err := userService.Repository.FindUserById(id)
	if err != nil {
		httphelper.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, user)
}

func (userService *UserService) FindOrCreateFromMicrosoftAuth(
	msUserInfo *MicrosoftUserInfo,
) (*UserEntity, error) {
	user, err := userService.Repository.FindUserByEmail(msUserInfo.Email)
	if err == nil {
		return user, nil
	}

	userToBeInserted := UserEntity{
		FirstName: msUserInfo.GivenName,
		LastName:  msUserInfo.FamilyName,
		Email:     msUserInfo.Email,
	}

	insertedUser, err := userService.Repository.InsertUser(&userToBeInserted)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from Microsoft auth: %w", err)
	}

	return insertedUser, nil
}

func (userService *UserService) HasAccess(userID string, permission string) (bool, error) {
	return userService.Repository.HasAccess(userID, permission)
}

func (userService *UserService) GenerateUserToken(user *UserEntity) (string, error) {
	expiration := time.Second * time.Duration(config.Variables.JwtExpirationInSeconds)

	roles, err := userService.Repository.GetRoles(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get user roles: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    user.ID,
		"email":     user.Email,
		"roles":     roles,
		"expiresAt": time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.Variables.JwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
