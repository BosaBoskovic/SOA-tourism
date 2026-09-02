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
	"stakeholders/rpc"
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

// seedAdmin creates the initial admin account on first startup only. It's a
// no-op (not an error) on every later restart, since ADMIN_USERNAME/EMAIL
// will already exist by then.
func seedAdmin(accountRepo *repo.AccountRepo) {
	ctx := context.Background()

	username := getEnvOrDefault("ADMIN_USERNAME", "admin")
	email := getEnvOrDefault("ADMIN_EMAIL", "admin@gmail.com")
	password := getEnvOrDefault("ADMIN_PASSWORD", "admin123")

	exists, err := accountRepo.ExistsByUsernameOrEmail(ctx, username, email)
	if err != nil {
		log.Printf("admin seed: could not check for existing admin account: %v", err)
		return
	}
	if exists {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("admin seed: could not hash admin password: %v", err)
		return
	}

	admin := model.Account{
		Username:     username,
		Email:        email,
		Role:         "admin",
		IsBlocked:    false,
		PasswordHash: string(hash),
	}

	if err := accountRepo.CreateAccount(ctx, admin); err != nil {
		log.Printf("admin seed: could not create admin account: %v", err)
		return
	}
	log.Printf("admin seed: created initial admin account %q", username)
}

func main() {
	logger := newLogger()
	ctx := context.Background()

	tp, err := initTracer(ctx)
	if err != nil {
		logger.Error("opentelemetry init failed", "error", err)
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("opentelemetry shutdown failed", "error", err)
		}
	}()

	neo4jURI := getEnvOrDefault("NEO4J_URI", "neo4j://localhost:7687")
	neo4jUser := getEnvOrDefault("NEO4J_USER", "neo4j")
	neo4jPassword := getEnvOrDefault("NEO4J_PASSWORD", "password")
	neo4jDatabase := getEnvOrDefault("NEO4J_DATABASE", "neo4j")
	jwtSecretRaw := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecretRaw == "" {
		log.Fatal("JWT_SECRET is not set; refusing to start with no signing secret (see .env.example)")
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

	r := gin.New()
	r.Use(gin.Recovery(), tracingMiddleware(), observabilityMiddleware(logger))
	grpcPort := getEnvOrDefault("STAKEHOLDERS_GRPC_PORT", "9091")
	go func() {
		if err := rpc.StartGRPCServer(grpcPort, authService, profileService); err != nil {
			log.Fatalf("neuspesno pokretanje gRPC servera: %v", err)
		}
	}()

	r.GET("/stakeholders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Stakeholders service radi"})
	})
	r.GET("/health", healthHandler(driver))
	r.GET("/metrics", metricsHandler())
	authHandler.RegisterRoutes(r)
	profileHandler.RegisterRoutes(r)

	logger.Info("stakeholders service starting", "port", 8081, "grpc_port", grpcPort)
	r.Run(":8081")
}
