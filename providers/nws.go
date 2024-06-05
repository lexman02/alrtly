package providers

import (
	"alrtly/webhook"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type NWS struct{}

type NWSResponse struct {
	Features []struct {
		Properties struct {
			ID        string `json:"id"`
			Event     string `json:"event"`
			Severity  string `json:"severity"`
			Urgency   string `json:"urgency"`
			Headline  string `json:"headline"`
			Response  string `json:"response"`
			Sent      string `json:"sent"`
			Effective string `json:"effective"`
			Expires   string `json:"expires"`
		} `json:"properties"`
	} `json:"features"`
}

type Alert struct {
	ID        string `json:"id"`
	Event     string `json:"event"`
	Severity  string `json:"severity"`
	Urgency   string `json:"urgency"`
	Headline  string `json:"headline"`
	Response  string `json:"response"`
	Sent      string `json:"sent"`
	Effective string `json:"effective"`
	Expires   string `json:"expires"`
}

func (n *NWS) FetchData() (interface{}, error) {
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
	url := fmt.Sprintf("https://api.weather.gov/alerts?status=actual,system&message_type=alert?point=%s,%s", lat, lon)
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
			ID:        feature.Properties.ID,
			Event:     feature.Properties.Event,
			Severity:  feature.Properties.Severity,
			Urgency:   feature.Properties.Urgency,
			Headline:  feature.Properties.Headline,
			Response:  feature.Properties.Response,
			Sent:      feature.Properties.Sent,
			Effective: feature.Properties.Effective,
			Expires:   feature.Properties.Expires,
		}

		alertData = append(alertData, alert)
	}

	fmt.Println("Fetching data from NWS...")

	return alertData, nil
}

func (n *NWS) PrepareData(data interface{}) (*webhook.WebhookData, error) {
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

	// Prepare the data for the webhook
	return &webhook.WebhookData{
		ID:       alerts[0].ID,
		Title:    alerts[0].Event,
		Content:  alerts[0].Headline,
		Priority: priority,
		Source:   "nws",
	}, nil
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
