package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func main() {
	ctx := context.Background()
	neo4jURI := getEnvOrDefault("NEO4J_URI", "neo4j://localhost:7687")
	neo4jUser := getEnvOrDefault("NEO4J_USER", "neo4j")
	neo4jPassword := getEnvOrDefault("NEO4J_PASSWORD", "password")
	neo4jDatabase := getEnvOrDefault("NEO4J_DATABASE", "neo4j")
	jwtSecretRaw := getEnvOrDefault("JWT_SECRET", "change-this-secret-in-production")
	if jwtSecretRaw == "change-this-secret-in-production" {
		log.Println("warning: JWT_SECRET is using a default value; set JWT_SECRET in production")
	}
	jwtSecret := []byte(jwtSecretRaw)

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("cannot create neo4j driver: %v", err)
	}
	defer func() {
		if closeErr := driver.Close(ctx); closeErr != nil {
			log.Printf("cannot close neo4j driver: %v", closeErr)
		}
	}()

	if err = driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("cannot connect to neo4j: %v", err)
	}

	authService := NewAuthService(driver, neo4jDatabase, jwtSecret)
	if err = authService.EnsureUniqueConstraints(ctx); err != nil {
		log.Fatalf("cannot create neo4j constraints: %v", err)
	}


	r := gin.Default()

	r.GET("/stakeholders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Stakeholders service radi",
		})
	})

	authService.RegisterRoutes(r)

	r.Run(":8081")
}