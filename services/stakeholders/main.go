package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"stakeholders/handler"
	"stakeholders/repo"
	"stakeholders/service"
)

import "golang.org/x/crypto/bcrypt"
import "stakeholders/model"

func getEnvOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func seedAdmin(accountRepo *repo.AccountRepo) {
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	admin := model.Account{
		Username:     "admin",
		Email:        "admin@gmail.com",
		Role:         "admin",
		IsBlocked:    false,
		PasswordHash: string(hash),
	}

	_ = accountRepo.CreateAccount(ctx, admin)
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

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("cannot create neo4j driver: %v", err)
	}
	defer func() {
		if err := driver.Close(ctx); err != nil {
			log.Printf("cannot close neo4j driver: %v", err)
		}
	}()

	if err = driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("cannot connect to neo4j: %v", err)
	}

	profileRepo := repo.NewProfileRepo(driver)

	accountRepo := repo.NewAccountRepo(driver, neo4jDatabase)
	seedAdmin(accountRepo)
	authService := service.NewAuthService(accountRepo, profileRepo, []byte(jwtSecretRaw))
	authHandler := handler.NewAuthHandler(authService)

	profileService := service.NewProfileService(profileRepo)
	profileHandler := handler.NewProfileHandler(profileService, authService)

	if err = authService.EnsureUniqueConstraints(ctx); err != nil {
		log.Fatalf("cannot create neo4j constraints: %v", err)
	}

	r := gin.Default()
	r.GET("/stakeholders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Stakeholders service radi"})
	})
	authHandler.RegisterRoutes(r)
	profileHandler.RegisterRoutes(r)

	r.Run(":8081")
}
