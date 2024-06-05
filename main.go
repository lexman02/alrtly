package main

import (
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

	"github.com/joho/godotenv"
)

var webhookURL string

func main() {
	fmt.Println("Loading .env file...")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	webhookURL = os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("WEBHOOK_URL is not set in the .env file")
	}

	r := server.NewRouter()
	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		fmt.Println("Starting server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Create a new goroutine to fetch data from the NWS provider
	go func() {
		// Create a new ticker that fires every 60 seconds
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		// Create a new NWS provider
		nws := &providers.NWS{}

		// Create a map to store sent alerts
		sentAlerts := make(map[string]webhook.WebhookData)

		// Fetch data from the NWS provider immediately
		err := fetchNWSData(nws, sentAlerts)
		if err != nil {
			log.Println(err)
		}

		// Fetch data from the NWS provider every 60 seconds
		for range ticker.C {
			err := fetchNWSData(nws, sentAlerts)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()

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

func fetchNWSData(nws *providers.NWS, sentAlerts map[string]webhook.WebhookData) error {
	// Fetch data from the NWS provider
	data, err := nws.FetchData()
	if err != nil {
		return err
	}

	if len(data.([]providers.Alert)) > 0 {
		// Prepare the data
		preparedData, err := nws.PrepareData(data)
		if err != nil {
			return err
		}

		if _, ok := sentAlerts[preparedData.ID]; ok {
			// Send the data to the webhook
			err := webhook.Send(webhookURL, *preparedData)
			if err != nil {
				return err
			}

			sentAlerts[preparedData.ID] = *preparedData
		}
	}

	return nil
}
