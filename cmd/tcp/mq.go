package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/diegodario88/sesamo/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type Message[T any] struct {
	DeliveryID int                    `json:"delivery_id"`
	RoutingKey string                 `json:"routing_key"`
	Body       T                      `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
	RawPayload []byte                 `json:"-"`
}

type MessageConsumer[T any] interface {
	Process(message *Message[T]) (int, error)
}

type ConsumerWrapper interface {
	ProcessNotification(notification *pgconn.Notification) (int, error)
}

type ConsumerWrapperImpl[T any] struct {
	consumer MessageConsumer[T]
}

func (cw ConsumerWrapperImpl[T]) ProcessNotification(
	notification *pgconn.Notification,
) (int, error) {
	msg, err := parseNotification[T](notification)
	if err != nil {
		return 0, err
	}
	return cw.consumer.Process(msg)
}

func WrapConsumer[T any](consumer MessageConsumer[T]) ConsumerWrapper {
	return ConsumerWrapperImpl[T]{consumer: consumer}
}

type MqListener struct {
	connectionString string
	consumers        map[string]ConsumerWrapper
	channelToQueue   map[string]string
	cancelFunc       context.CancelFunc
	storage          *sqlx.DB // Referência ao storage principal para operações de ACK/NACK
}

func NewMqListener(storage *sqlx.DB) *MqListener {
	return &MqListener{
		connectionString: config.Variables.DatabaseUrl,
		consumers:        make(map[string]ConsumerWrapper),
		channelToQueue:   make(map[string]string),
		storage:          storage,
	}
}

func (mq *MqListener) RegisterConsumer(queue string, consumer ConsumerWrapper) *MqListener {
	mq.consumers[queue] = consumer
	return mq
}

func (mq *MqListener) ListenForNotifications(ctx context.Context) error {
	var innerCtx context.Context
	innerCtx, mq.cancelFunc = context.WithCancel(ctx)

	notifyConn, err := pgx.Connect(ctx, mq.connectionString)
	if err != nil {
		return err
	}

	defer func() {
		log.Println("MQ listener main loop exited")
	}()

	if len(mq.consumers) == 0 {
		return fmt.Errorf("No consumers registered, nothing to listen for")
	}

	for queue := range mq.consumers {
		var channelId string
		qChannel := fmt.Sprintf("SELECT mq.open_channel('%s', 1)", queue)
		err = notifyConn.QueryRow(ctx, qChannel).Scan(&channelId)

		if err != nil {
			return fmt.Errorf("Error opening channel for %s: %v\n", queue, err)
		}

		mq.channelToQueue[channelId] = queue

		log.Printf(
			"Listening for notifications on queue: %s via channel ID: %s\n",
			queue,
			channelId,
		)
	}

	log.Println("PostgreSQL notification listener started")

	var currentDeliveryID int
	var processingMessage atomic.Bool

	for {
		select {
		case <-innerCtx.Done():
			log.Println("MQ listener shutdown signal received")

			if processingMessage.Load() && currentDeliveryID > 0 {
				log.Printf("Sending NACK for in-progress message ID: %d", currentDeliveryID)
				if err := mq.disacknowledgeMessage(currentDeliveryID, "5 minutes"); err != nil {
					log.Printf("Error during shutdown NACK: %v", err)
				}
			}

			return innerCtx.Err()
		default:
			notification, err := notifyConn.WaitForNotification(ctx)
			if err != nil {
				return err
			}

			log.Printf("Received notification on channel %s\n", notification.Channel)

			consumer, exists := mq.consumers[mq.channelToQueue[notification.Channel]]
			if !exists {
				log.Printf(
					"No consumer registered for channel %s, skipping\n",
					notification.Channel,
				)
				continue
			}

			processingMessage.Store(true)
			deliveryId, err := consumer.ProcessNotification(notification)
			currentDeliveryID = deliveryId

			if err != nil {
				log.Printf("Error processing message on %s: %v\n", notification.Channel, err)

				if deliveryId > 0 {
					log.Printf("Sending NACK for failed processing message ID: %d", deliveryId)
					if err := mq.disacknowledgeMessage(currentDeliveryID, "5 minutes"); err != nil {
						log.Printf("Error during shutdown NACK: %v", err)
					}
				}

				processingMessage.Store(false)
				continue
			}

			if err := mq.acknowledgeMessage(currentDeliveryID); err != nil {
				log.Printf("Error acknowledging message: %v\n", err)
				if nackErr := mq.disacknowledgeMessage(currentDeliveryID, "1 minute"); nackErr != nil {
					log.Printf("Error sending NACK after failed ACK: %v", nackErr)
				}
			}

			processingMessage.Store(false)
		}
	}
}

func (mq *MqListener) Shutdown(ctx context.Context, conn *sqlx.DB) error {
	if mq.cancelFunc == nil {
		log.Println("MqListener is not running or is already canceled")

		return nil
	}

	log.Println("Gracefully shutting down MQ listener...")
	mq.cancelFunc()
	time.Sleep(100 * time.Millisecond)

	log.Println("Closing channels...")
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for c := range mq.channelToQueue {
		q := fmt.Sprintf("CALL mq.close_channel(%s);", c)
		_, err := conn.ExecContext(cleanupCtx, q)
		if err != nil {
			log.Printf("Error closing channel: %v\n", err)
		}

		_, err = conn.ExecContext(cleanupCtx, "CALL mq.close_dead_channels();")
		if err != nil {
			log.Printf("Error closing dead channels: %v\n", err)
		}
	}

	mq.channelToQueue = make(map[string]string)
	log.Println("MQ listener cleanup process completed")
	return nil
}

func parseNotification[T any](notification *pgconn.Notification) (*Message[T], error) {
	var temp struct {
		DeliveryID int                    `json:"delivery_id"`
		RoutingKey string                 `json:"routing_key"`
		Headers    map[string]interface{} `json:"headers"`
		Body       json.RawMessage        `json:"body"`
	}

	if err := json.Unmarshal([]byte(notification.Payload), &temp); err != nil {
		return nil, fmt.Errorf("error parsing message payload: %w", err)
	}

	var body T
	if err := json.Unmarshal(temp.Body, &body); err != nil {
		return nil, fmt.Errorf("error unmarshaling message body: %w", err)
	}

	return &Message[T]{
		DeliveryID: temp.DeliveryID,
		RoutingKey: temp.RoutingKey,
		Headers:    temp.Headers,
		Body:       body,
		RawPayload: []byte(notification.Payload),
	}, nil
}

func (mq *MqListener) acknowledgeMessage(deliveryID int) error {
	_, err := mq.storage.Exec(fmt.Sprintf("CALL mq.ack(%d)", deliveryID))
	if err != nil {
		return fmt.Errorf("error sending ACK message: %w", err)
	}

	log.Println("Message ACK successfully")
	return nil
}

func (mq *MqListener) disacknowledgeMessage(deliveryID int, retryAfter string) error {
	_, err := mq.storage.Exec("CALL mq.nack($1,retry_after=>$2)", deliveryID, retryAfter)
	if err != nil {
		return fmt.Errorf("error sending NACK message:: %w", err)
	}

	log.Println("Message NACK successfully")
	return nil
}
