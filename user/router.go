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

	protected.Handle("/users", RBACMiddleware(h, "users:read")(
		http.HandlerFunc(h.GetAllUsers))).Methods("GET")

	protected.Handle("/users/{id}", RBACMiddleware(h, "users:read")(
		http.HandlerFunc(h.GetUserByID))).Methods("GET")

	return h
}
