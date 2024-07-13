package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/diegodario88/sesamo/httphelper"
	"github.com/diegodario88/sesamo/user"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type APIServer struct {
	port int64
	db   *sqlx.DB
}

type Info struct {
	Service   string
	Condition string
}

type Alive struct {
	Status string
	Info   Info
}

func NewServer(port int64, db *sqlx.DB) *APIServer {
	return &APIServer{
		port: port,
		db:   db,
	}
}

func (api APIServer) Run() {
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

	log.Printf("API Server listening at http://localhost:%d", api.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", api.port), router)

	if err != nil {
		log.Fatal(err)
	}
}
