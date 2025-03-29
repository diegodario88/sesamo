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
	GetUserOrganizations(userId string) ([]OrganizationEntity, error)
	GetUserBranches(userId string, orgId string) ([]BranchEntity, error)
	FindOrganizationUsers(orgID string) ([]UserEntity, error)
	FindOrganizationUserByID(orgID string, userID string) (*UserEntity, error)
	FindBranchUsers(orgID string, branchID string) ([]UserEntity, error)
	GetOrganizationBranches(orgID string) ([]BranchEntity, error)
}

type UserService struct {
	Repo IUserRepository
}

func NewUserService(db *sqlx.DB) UserService {
	var newUserService = UserService{
		Repo: NewUserRepository(db),
	}

	return newUserService
}

func (svc *UserService) Login(w http.ResponseWriter, r *http.Request) {
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

	user, err := svc.authenticateUserByEmailPassword(loginUserPayload)

	if err != nil {
		httphelper.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("not found, invalid email or password"),
		)
		return
	}

	token, err := svc.GenerateUserToken(user)

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (svc *UserService) Register(w http.ResponseWriter, r *http.Request) {
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

	user, err := svc.Repo.FindUserByEmail(registerUserPayload.Email)

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

	insertedUser, err := svc.Repo.InsertUser(&userToBeInserted)

	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusCreated, insertedUser)
}

func (svc *UserService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := svc.Repo.FindAllUsers()
	if err != nil {
		httphelper.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, users)
}

func (svc *UserService) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		httphelper.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	user, err := svc.Repo.FindUserById(id)
	if err != nil {
		httphelper.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, user)
}

func (svc *UserService) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	user, err := svc.Repo.FindUserById(userID)
	if err != nil {
		httphelper.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, user)
}

func (svc *UserService) FindUserOrganizations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	orgs, err := svc.Repo.GetUserOrganizations(userID)
	if err != nil {
		httphelper.WriteError(w, http.StatusNotFound, fmt.Errorf("orgs not found"))
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, orgs)
}

func (svc *UserService) FindOrCreateFromMicrosoftAuth(
	msUserInfo *MicrosoftUserInfo,
) (*UserEntity, error) {
	user, err := svc.Repo.FindUserByEmail(msUserInfo.Email)
	if err == nil {
		return user, nil
	}

	userToBeInserted := UserEntity{
		FirstName: msUserInfo.GivenName,
		LastName:  msUserInfo.FamilyName,
		Email:     msUserInfo.Email,
	}

	insertedUser, err := svc.Repo.InsertUser(&userToBeInserted)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from Microsoft auth: %w", err)
	}

	return insertedUser, nil
}

func (svc *UserService) HasAccess(userID string, permission string) (bool, error) {
	return svc.Repo.HasAccess(userID, permission)
}

func (svc *UserService) FindUserBranches(userID string, orgId string) ([]BranchEntity, error) {
	return svc.Repo.GetUserBranches(userID, orgId)
}

func (svc *UserService) GenerateUserToken(user *UserEntity) (string, error) {
	expiration := time.Second * time.Duration(config.Variables.JwtExpirationInSeconds)

	roles, err := svc.Repo.GetRoles(user.ID)
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

func (svc *UserService) GetOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]

	users, err := svc.Repo.FindOrganizationUsers(orgID)
	if err != nil {
		http.Error(w, "Error retrieving organization users", http.StatusInternalServerError)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, users)
}

func (svc *UserService) GetOrganizationUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	userID := vars["id"]

	user, err := svc.Repo.FindOrganizationUserByID(orgID, userID)
	if err != nil {
		http.Error(w, "User not found or not in this organization", http.StatusNotFound)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, user)
}

func (svc *UserService) GetBranchUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	branchID := vars["branchId"]

	users, err := svc.Repo.FindBranchUsers(orgID, branchID)
	if err != nil {
		http.Error(w, "Error retrieving branch users", http.StatusInternalServerError)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, users)
}

func (svc *UserService) GetOrganizationBranches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]

	branches, err := svc.Repo.GetOrganizationBranches(orgID)
	if err != nil {
		httphelper.WriteError(
			w,
			http.StatusInternalServerError,
			fmt.Errorf("error retrieving branches: %w", err),
		)
		return
	}

	httphelper.WriteJSON(w, http.StatusOK, branches)
}

func (svc *UserService) authenticateUserByEmailPassword(
	loginUserPayload LoginUserPayload,
) (*UserEntity, error) {
	user, err := svc.Repo.FindUserByEmail(loginUserPayload.Email)

	if err != nil {
		return nil, err
	}

	if _, err := user.CheckPassword(loginUserPayload.Password); err != nil {
		return nil, err
	}

	return user, nil
}
