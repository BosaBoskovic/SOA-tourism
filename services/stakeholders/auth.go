package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/crypto/bcrypt"
)

const accessTokenTTL = 15 * time.Minute

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Email    string `json:"email" binding:"required,email,max=120"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail" binding:"required,min=3,max=120"`
	Password        string `json:"password" binding:"required,min=8,max=100"`
}

type AccountResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type accountRecord struct {
	Username     string
	Email        string
	Role         string
	PasswordHash string
}

type AccessClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	driver   neo4j.DriverWithContext
	database string
	secret   []byte
}

func NewAuthService(driver neo4j.DriverWithContext, database string, secret []byte) *AuthService {
	return &AuthService{
		driver:   driver,
		database: database,
		secret:   secret,
	}
}

func (s *AuthService) RegisterRoutes(r *gin.Engine) {
	r.POST("/stakeholders/register", s.register)
	r.POST("/stakeholders/login", s.login)
}

func (s *AuthService) EnsureUniqueConstraints(ctx context.Context) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.database})
	defer func() {
		if closeErr := session.Close(ctx); closeErr != nil {
			log.Printf("cannot close neo4j session while creating constraints: %v", closeErr)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		queries := []string{
			`CREATE CONSTRAINT account_username_unique IF NOT EXISTS
			 FOR (u:Account)
			 REQUIRE u.username IS UNIQUE`,
			`CREATE CONSTRAINT account_email_unique IF NOT EXISTS
			 FOR (u:Account)
			 REQUIRE u.email IS UNIQUE`,
		}

		for _, query := range queries {
			if _, runErr := tx.Run(ctx, query, nil); runErr != nil {
				return nil, runErr
			}
		}

		return nil, nil
	})

	return err
}

func normalizeRegistrableRole(role string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "guide", "vodic":
		return "guide", nil
	case "tourist", "turista":
		return "tourist", nil
	default:
		return "", errors.New("role must be guide or tourist")
	}
}

func generateAccessToken(secret []byte, account accountRecord) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(accessTokenTTL)

	claims := AccessClaims{
		Role:  account.Role,
		Email: account.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *AuthService) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := normalizeRegistrableRole(req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dozvoljene uloge za registraciju su vodic i turista"})
		return
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Neuspesno hesiranje lozinke"})
		return
	}

	ctx := c.Request.Context()
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.database})
	defer func() {
		if closeErr := session.Close(ctx); closeErr != nil {
			log.Printf("cannot close neo4j session: %v", closeErr)
		}
	}()

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		existsResult, runErr := tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($username) OR toLower(u.email) = toLower($email)
			 RETURN count(u) AS total`,
			map[string]any{
				"username": req.Username,
				"email":    req.Email,
			},
		)
		if runErr != nil {
			return nil, runErr
		}

		record, singleErr := existsResult.Single(ctx)
		if singleErr != nil {
			return nil, singleErr
		}

		total, _ := record.Get("total")
		if count, ok := total.(int64); ok && count > 0 {
			return nil, errors.New("username_or_email_exists")
		}

		_, runErr = tx.Run(ctx,
			`CREATE (u:Account {
				username: $username,
				passwordHash: $passwordHash,
				email: $email,
				role: $role,
				createdAt: datetime()
			})`,
			map[string]any{
				"username":     req.Username,
				"passwordHash": string(passwordHashBytes),
				"email":        req.Email,
				"role":         role,
			},
		)

		return nil, runErr
	})
	if err != nil {
		if err.Error() == "username_or_email_exists" || strings.Contains(strings.ToLower(err.Error()), "constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Korisnicko ime ili email vec postoje"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri cuvanju naloga"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Uspesna registracija",
		"account": AccountResponse{
			Username:  req.Username,
			Email:     req.Email,
			Role:      role,
			CreatedAt: time.Now().UTC(),
		},
	})
}

func (s *AuthService) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.database})
	defer func() {
		if closeErr := session.Close(ctx); closeErr != nil {
			log.Printf("cannot close neo4j session: %v", closeErr)
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, runErr := tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($identity) OR toLower(u.email) = toLower($identity)
			 RETURN u.username AS username, u.email AS email, u.role AS role, u.passwordHash AS passwordHash
			 LIMIT 1`,
			map[string]any{"identity": req.UsernameOrEmail},
		)
		if runErr != nil {
			return nil, runErr
		}

		records, collectErr := res.Collect(ctx)
		if collectErr != nil {
			return nil, collectErr
		}
		if len(records) == 0 {
			return nil, errors.New("invalid_credentials")
		}

		record := records[0]
		username, _ := record.Get("username")
		email, _ := record.Get("email")
		role, _ := record.Get("role")
		passwordHash, _ := record.Get("passwordHash")

		return accountRecord{
			Username:     username.(string),
			Email:        email.(string),
			Role:         role.(string),
			PasswordHash: passwordHash.(string),
		}, nil
	})
	if err != nil {
		if err.Error() == "invalid_credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Pogresni kredencijali"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri prijavi"})
		return
	}

	account := result.(accountRecord)
	if compareErr := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); compareErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Pogresni kredencijali"})
		return
	}

	token, expiresAt, tokenErr := generateAccessToken(s.secret, account)
	if tokenErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri generisanju tokena"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Uspesna prijava",
		"accessToken": token,
		"tokenType":   "Bearer",
		"expiresIn":   int64(accessTokenTTL.Seconds()),
		"expiresAt":   expiresAt.Format(time.RFC3339),
		"account": gin.H{
			"username": account.Username,
			"email":    account.Email,
			"role":     account.Role,
		},
	})
}
