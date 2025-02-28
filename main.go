package main

import (
	"alrtly/config"
	"alrtly/providers"
	"alrtly/server"
	"alrtly/webhook"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var webhookURL string

func main() {
	fmt.Println("Loading .env file...")
	if err := config.Init(); err != nil {
		log.Fatal("failed to load config:", err)
	}

	cfg := config.Get()
	webhookClient := webhook.NewClient(cfg.WebhookURL)

	// Initialize the server
	r := server.NewRouter(cfg)
	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		fmt.Println("Starting server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Start polling for each provider
	nws := &providers.NWS{}
	go nws.Poll(webhookClient, 60*time.Second)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
