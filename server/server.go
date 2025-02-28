package server

import (
	"alrtly/config"
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// go:embed audio/*
var audioFS embed.FS

func NewRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	h := NewHandler(cfg)

	r.StaticFS("/audio", http.FS(audioFS))
	r.POST("/alert", h.PostAlert)
	r.GET("/test/{provider}", h.TestAlert)

	return r
}
