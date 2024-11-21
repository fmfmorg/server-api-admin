package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server-api-admin/endpoints"
	"server-api-admin/util/postgresdb"
	"server-api-admin/util/redisclient"
	"server-api-admin/util/router"
)

func main() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())

	// Listen for termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start your services with the context
	go func() {
		log.Println("Starting services...")
		endpoints.Listen()
		router.Listen(ctx)
	}()

	// Wait for a termination signal
	<-signalChan
	log.Println("Received termination signal, shutting down...")

	// Cancel the context to signal services to shut down
	cancel()

	// Give services some time to shut down gracefully
	time.Sleep(5 * time.Second)

	// Close database and redis connections
	postgresdb.DB.Close()
	redisclient.Client.Close()

	log.Println("Shutdown complete")
}
