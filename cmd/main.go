package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/diegodario88/sesamo/cmd/http"
	mq "github.com/diegodario88/sesamo/cmd/tcp"
	"github.com/diegodario88/sesamo/db"
	"github.com/diegodario88/sesamo/user"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	storage, err := db.CreateStorageConn()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	userConsumer := mq.WrapConsumer(user.NewConsumer(storage))
	mqListener := mq.NewMqListener(storage).RegisterConsumer("create_user", userConsumer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting MQ listener...")
		if err := mqListener.ListenForNotifications(ctx); err != nil && err != context.Canceled {
			log.Printf("MQ listener error: %v", err)
		} else if err == context.Canceled {
			log.Println("MQ listener shut down successfully")
		}
	}()

	httpServer := api.NewServer(storage)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
		log.Println("HTTP server stopped")
	}()

	<-ctx.Done()
	log.Println("Shutdown signal  received. Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	startTime := time.Now()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Printf("HTTP server shutdown successful (took %v)", time.Since(startTime))
	}

	startTime = time.Now()
	if err := mqListener.Shutdown(shutdownCtx, storage); err != nil {
		log.Printf("MQ listener shutdown error: %v", err)
	} else {
		log.Printf("MQ listener shutdown successful (took %v)", time.Since(startTime))
	}

	waitCh := make(chan struct{})

	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		log.Println("All services stopped successfully")
	case <-time.After(15 * time.Second):
		log.Println("Shutdown timed out, some services may not have stopped gracefully")
	}

	log.Println("Closing storage ...")
	if err := storage.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Application shutdown completed")
}
