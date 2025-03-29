package user

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	UserService
}

func NewHandler(userService UserService) *Handler {
	return &Handler{userService}
}

func (h *Handler) RegisterRoutes(router *mux.Router) *Handler {
	router.HandleFunc("/users/login", h.Login).Methods("POST")
	router.HandleFunc("/users/register", h.Register).Methods("POST")

	protected := router.PathPrefix("/").Subrouter()
	protected.Use(AuthMiddleware)

	protected.HandleFunc("/users/me", h.GetCurrentUser).Methods("GET")
	protected.HandleFunc("/users/organizations", h.FindUserOrganizations).Methods("GET")

	orgRouter := protected.PathPrefix("/organizations/{orgId}").Subrouter()
	orgRouter.Use(h.OrganizationAccessMiddleware)

	orgRouter.Handle("/branches", RBACMiddleware(h, "branches:read")(
		http.HandlerFunc(h.GetOrganizationBranches))).Methods("GET")

	orgRouter.Handle("/users", RBACMiddleware(h, "users:read")(
		http.HandlerFunc(h.GetOrganizationUsers))).Methods("GET")

	orgRouter.Handle("/users/{id}", RBACMiddleware(h, "users:read")(
		http.HandlerFunc(h.GetOrganizationUserByID))).Methods("GET")

	branchRouter := orgRouter.PathPrefix("/branches/{branchId}").Subrouter()
	branchRouter.Use(h.BranchAccessMiddleware)

	branchRouter.Handle("/users", RBACMiddleware(h, "users:read")(
		http.HandlerFunc(h.GetBranchUsers))).Methods("GET")

	return h
}
