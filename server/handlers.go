package server

import (
	"alrtly/webhook"
	"os"

	"github.com/gin-gonic/gin"
)

func PostAlert(c *gin.Context) {
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
	webhookURL := os.Getenv("WEBHOOK_URL")
	if err := webhook.Send(webhookURL, data); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": "alert sent"})
}
