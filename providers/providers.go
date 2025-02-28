package providers

import (
	"alrtly/webhook"
	"time"
)

type Provider interface {
	FetchData() (interface{}, error)
	PrepareData(data interface{}) (*webhook.WebhookData, error)
	TestAlert(client *webhook.Client) error
	Poll(client *webhook.Client, interval time.Duration)
}

type ProviderFactory func() Provider

var providers = make(map[string]ProviderFactory)

func RegisterProvider(name string, factory ProviderFactory) {
	providers[name] = factory
}

func GetProvider(name string) (Provider, bool) {
	factory, exists := providers[name]
	if !exists {
		return nil, false
	}

	return factory(), true
}
