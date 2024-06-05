package server

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// go:embed audio/*
var audioFS embed.FS

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.StaticFS("/audio", http.FS(audioFS))
	r.POST("/alert", PostAlert)

	return r
}
