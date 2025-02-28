package server

import (
	"alrtly/config"
	"alrtly/providers"
	"alrtly/webhook"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	webhookClient *webhook.Client
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		webhookClient: webhook.NewClient(cfg.WebhookURL),
	}
}

func (h *Handler) PostAlert(c *gin.Context) {
	var data webhook.WebhookData

	// Handle the request
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Set the source of the data
	if data.Source != "api" {
		data.Source = "api"
	}

	// Process the data
	if err := h.webhookClient.Send(data); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": "alert sent"})
}

func (h *Handler) TestAlert(c *gin.Context) {
	provider, ok := c.Params.Get("provider")
	if !ok {
		c.JSON(400, gin.H{"error": "provider not specified"})
		return
	}

	p, exists := providers.GetProvider(provider)
	if !exists {
		c.JSON(400, gin.H{"error": "invalid provider"})
		return
	}

	// Fetch data from the provider
	if err := p.TestAlert(h.webhookClient); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": "test alert sent successfully"})
}
