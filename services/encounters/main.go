package main

import (
	"context"
	"log"
	"os"
	"time"

	"encounters/auth"
	"encounters/handler"
	"encounters/repository"
	"encounters/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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

	verifier, err := auth.NewVerifier()
	if err != nil {
		logger.Error("auth verifier init failed", "error", err)
		log.Fatal(err)
	}

	// Shares the tours MongoDB server (its own "encountersDB" database, so
	// data is still fully isolated) rather than standing up a whole new
	// database container for what's a modest feature.
	mongoURI := getEnvOrDefault("MONGODB_URI", "localhost:27017")
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://" + mongoURI + "/?connect=direct"))
	if err != nil {
		log.Fatal("MongoDB connect error:", err)
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}
	logger.Info("connected to MongoDB")
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			logger.Error("MongoDB disconnect error", "error", err)
		}
	}()

	db := client.Database("encountersDB")
	encounterRepo := repository.NewEncounterRepository(db)
	progressRepo := repository.NewProgressRepository(db)
	encounterService := service.NewEncounterService(encounterRepo, progressRepo)
	encounterHandler := handler.NewEncounterHandler(encounterService)

	r := gin.New()
	r.Use(gin.Recovery(), tracingMiddleware(), observabilityMiddleware(logger))

	r.GET("/health", healthHandler(client))
	r.GET("/metrics", metricsHandler())
	encounterHandler.RegisterRoutes(r, verifier)

	port := getEnvOrDefault("PORT", "8083")
	logger.Info("encounters service starting", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Error("encounters service stopped", "error", err)
	}
}
