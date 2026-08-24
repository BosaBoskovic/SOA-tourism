package main

import (
	"context"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := newLogger()

	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		logger.Error("opentelemetry init failed", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("opentelemetry shutdown failed", "error", err)
			}
		}()
	}

	r := gin.New()
	r.Use(gin.Recovery(), tracingMiddleware(), observabilityMiddleware(logger))

	r.GET("/encounters", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Encounters service radi",
		})
	})
	r.GET("/health", healthHandler())
	r.GET("/metrics", metricsHandler())

	logger.Info("encounters service starting", "port", 8083)
	if err := r.Run(":8083"); err != nil {
		logger.Error("encounters service stopped", "error", err)
	}
}
