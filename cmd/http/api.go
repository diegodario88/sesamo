package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/diegodario88/sesamo/config"
	"github.com/diegodario88/sesamo/httphelper"
	"github.com/diegodario88/sesamo/user"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type APIServer struct {
	port   int64
	db     *sqlx.DB
	server *http.Server
}

type Info struct {
	Service   string
	Condition string
}

type Alive struct {
	Status string
	Info   Info
}

func NewServer(db *sqlx.DB) *APIServer {
	return &APIServer{
		port: config.Variables.Port,
		db:   db,
	}
}

func (api *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	user.NewHandler(user.NewUserService(api.db)).RegisterRoutes(subrouter)

	liveness := func(w http.ResponseWriter, r *http.Request) {
		log.Println("HTTP Server is alive!")
		httphelper.WriteJSON(w, http.StatusOK, Alive{
			Status: "ok",
			Info:   Info{Service: "Sesamo", Condition: "up"},
		})
	}

	router.Handle("/live", http.HandlerFunc(liveness))

	api.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", api.port),
		Handler: router,
	}

	log.Printf("API Server listening at http://localhost:%d", api.port)

	err := api.server.ListenAndServe()
	if err != nil {
		return err
	}

	log.Printf("API Server listening at http://localhost:%d", api.port)
	return nil
}

func (api *APIServer) Shutdown(ctx context.Context) error {
	if api.server != nil {
		log.Println("Calling gracefully http shutdown...")
		return api.server.Shutdown(ctx)
	}
	return nil
}
