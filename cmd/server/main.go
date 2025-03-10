package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soa-rs/fit/internal/config"
	"github.com/soa-rs/fit/internal/config/logger"
)

func main() {
	config.LoadEnvs()
	config.SetupLogger()

	engine := gin.Default()
	engine.SetTrustedProxies(nil)

	// Health check route
	engine.GET("/health", func(c *gin.Context) {
		logger.LogTrace("Health check route")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start server
	// Sometimes, logs sent before initializing the logger are printed
	// before those sent after said initialization. This is because the
	// before-logs are processed in a separate goroutine, which may be
	// executed before or after the main thread.
	port := config.GetEnvOrDefault(config.EnvBackendPort)
	host := config.GetEnvOrDefault(config.EnvBackendHost)
	if err := engine.Run(fmt.Sprintf("%s:%s", host, port)); err != nil {
		logger.LogFatal("Failed to run server: %v", err)
	}
}
