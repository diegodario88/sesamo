package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/diegodario88/sesamo/utils"
)

type APIServer struct {
	port int64
	db   *sql.DB
}

type Info struct {
	Service   string
	Condition string
}

type Alive struct {
	Status string
	Info   Info
}

func NewServer(port int64, db *sql.DB) *APIServer {
	return &APIServer{
		port: port,
		db:   db,
	}
}

func (api APIServer) Run() {
	mux := http.NewServeMux()

	liveness := func(w http.ResponseWriter, r *http.Request) {
		log.Println("HTTP Server is alive!")
		utils.WriteJSON(w, http.StatusOK, Alive{
			Status: "ok",
			Info:   Info{Service: "Sesamo", Condition: "up"},
		})
	}

	mux.Handle("/live", http.HandlerFunc(liveness))

	log.Printf("API Server listening at http://localhost:%d", api.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", api.port), mux)

	if err != nil {
		log.Fatal(err)
	}
}
