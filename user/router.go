package user

import "github.com/gorilla/mux"

type Handler struct {
	UserService
}

func NewHandler(userService UserService) *Handler {
	return &Handler{userService}
}

func (h *Handler) RegisterRoutes(router *mux.Router) *Handler {
	router.HandleFunc("/users/login", h.Login).Methods("POST")
	router.HandleFunc("/users/register", h.Register).Methods("POST")
	return h
}
