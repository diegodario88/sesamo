package user

import (
	"log"

	mq "github.com/diegodario88/sesamo/cmd/tcp"
	"github.com/jmoiron/sqlx"
)

type Consumer struct {
	UserService
}

type NewUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewConsumer(storage *sqlx.DB) *Consumer {
	userService := NewUserService(storage)
	return &Consumer{userService}
}

func (consumer *Consumer) Process(message *mq.Message[NewUser]) (int, error) {
	//TODO: Aqui a gente pode usar o service para de fato persistir o user

	log.Println("Processing message for user %s", message.Body.LastName)
	return message.DeliveryID, nil
}
