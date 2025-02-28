package providers

import (
	"alrtly/webhook"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type NWS struct{}

type NWSResponse struct {
	Features []struct {
		Properties struct {
			ID         string `json:"id"`
			Event      string `json:"event"`
			SenderName string `json:"senderName"`
			Severity   string `json:"severity"`
			Urgency    string `json:"urgency"`
			Headline   string `json:"headline"`
			Response   string `json:"response"`
			Sent       string `json:"sent"`
			Effective  string `json:"effective"`
			Expires    string `json:"expires"`
		} `json:"properties"`
	} `json:"features"`
}

type Alert struct {
	ID         string `json:"id"`
	Event      string `json:"event"`
	SenderName string `json:"senderName"`
	Severity   string `json:"severity"`
	Urgency    string `json:"urgency"`
	Headline   string `json:"headline"`
	Response   string `json:"response"`
	Sent       string `json:"sent"`
	Effective  string `json:"effective"`
	Expires    string `json:"expires"`
}

func init() {
	RegisterProvider("nws", func() Provider {
		return &NWS{}
	})
}

func (n NWS) FetchData() (interface{}, error) {
	// Get the coordinates of the user's location from the environment variables
	lat := os.Getenv("LATITUDE")
	lon := os.Getenv("LONGITUDE")

	if lat == "" || lon == "" {
		var err error

		// Get the coordinates from the address
		lat, lon, err = getCoordinates()
		if err != nil {
			return nil, errors.New("coordinates not found")
		}

		// Set the coordinates in the environment variables
		os.Setenv("LATITUDE", lat)
		os.Setenv("LONGITUDE", lon)
	}

	// Fetch alert data
	url := fmt.Sprintf("https://api.weather.gov/alerts?status=actual,system&message_type=alert&point=%s,%s", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract the alert data
	var alerts NWSResponse
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		return nil, err
	}

	var alertData []Alert
	for _, feature := range alerts.Features {
		alert := Alert{
			ID:         feature.Properties.ID,
			Event:      feature.Properties.Event,
			SenderName: feature.Properties.SenderName,
			Severity:   feature.Properties.Severity,
			Urgency:    feature.Properties.Urgency,
			Headline:   feature.Properties.Headline,
			Response:   feature.Properties.Response,
			Sent:       feature.Properties.Sent,
			Effective:  feature.Properties.Effective,
			Expires:    feature.Properties.Expires,
		}

		alertData = append(alertData, alert)
	}

	return alertData, nil
}

func (n NWS) PrepareData(data interface{}) (*webhook.WebhookData, error) {
	// Prepare the data for the webhook
	alerts, ok := data.([]Alert)
	if !ok {
		return nil, errors.New("invalid data type")
	}

	// Parse the event priority
	var priority string
	if strings.Contains(alerts[0].Severity, "Extreme") || strings.Contains(alerts[0].Urgency, "Immediate") || strings.Contains(alerts[0].Event, "Warning") {
		priority = "high"
	} else if strings.Contains(alerts[0].Event, "Watch") {
		priority = "medium"
	} else {
		priority = "low"
	}

	// Add extra grammar in headline for better readability
	alerts[0].Headline = "A " + alerts[0].Headline
	alerts[0].Headline = strings.ReplaceAll(alerts[0].Headline, " issued", " has been issued on")
	// Expand NWS to National Weather Service in headline
	alerts[0].Headline = strings.ReplaceAll(alerts[0].Headline, "NWS", "the National Weather Service in")

	// Prepare the data for the webhook
	return &webhook.WebhookData{
		ID:       alerts[0].ID,
		Title:    alerts[0].Event,
		Content:  alerts[0].Headline,
		Priority: priority,
		Source:   "nws",
	}, nil
}

func (n NWS) TestAlert(client *webhook.Client) error {
	// Implement the method to send a test alert
	// This could involve calling the PostAlert function with some test data
	testData := Alert{
		ID:       "test-id",
		Event:    "Test Event",
		Headline: "This is a test alert",
		Severity: "Minor",
		Urgency:  "Future",
		// Fill in the fields of the Alert struct with test data
	}

	alertData := []Alert{testData}
	webhookData, err := n.PrepareData(alertData)
	if err != nil {
		return err
	}

	return client.Send(*webhookData)
}

func (n NWS) Poll(client *webhook.Client, interval time.Duration) {
	// Create a map to store sent alerts
	sentAlerts := make(map[string]webhook.WebhookData)

	// Create a new ticker that fires at the specified interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Poll immediately on start
	if err := n.pollOnce(client, sentAlerts); err != nil {
		log.Printf("NWS polling error: %v", err)
	}

	// Poll on ticker interval
	for range ticker.C {
		if err := n.pollOnce(client, sentAlerts); err != nil {
			log.Printf("NWS polling error: %v", err)
			continue
		}
	}
}

func (n NWS) pollOnce(client *webhook.Client, sentAlerts map[string]webhook.WebhookData) error {
	// Fetch data from the NWS provider
	data, err := n.FetchData()
	if err != nil {
		return err
	}

	if len(data.([]Alert)) > 0 {
		// Prepare the data
		preparedData, err := n.PrepareData(data)
		if err != nil {
			return err
		}

		if _, ok := sentAlerts[preparedData.ID]; !ok {
			// Send the data to the webhook
			err := client.Send(*preparedData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getCoordinates() (string, string, error) {
	type address struct {
		Street string
		City   string
		State  string
	}

	// Get the user's address from the environment variables
	addr := &address{
		Street: os.Getenv("STREET"),
		City:   os.Getenv("CITY"),
		State:  os.Getenv("STATE"),
	}

	if addr.Street == "" || addr.City == "" || addr.State == "" {
		return "", "", errors.New("address not found")
	}

	// Geocode the address to get the coordinates
	url := fmt.Sprintf("https://geocoding.geo.census.gov/geocoder/locations/address?street=%s&city=%s&state=%s&benchmark=2020&format=json", strings.ReplaceAll(addr.Street, " ", "+"), strings.ReplaceAll(addr.City, " ", "+"), addr.State)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Extract the coordinates
	var coordinates struct {
		Result struct {
			AddressMatches []struct {
				Coordinates struct {
					Latitude  float64 `json:"y"`
					Longitude float64 `json:"x"`
				} `json:"coordinates"`
			} `json:"addressMatches"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &coordinates)
	if err != nil {
		return "", "", err
	}

	if len(coordinates.Result.AddressMatches) == 0 {
		return "", "", errors.New("no coordinates found")
	}

	lat := fmt.Sprintf("%f", coordinates.Result.AddressMatches[0].Coordinates.Latitude)
	lon := fmt.Sprintf("%f", coordinates.Result.AddressMatches[0].Coordinates.Longitude)

	return lat, lon, nil
}
