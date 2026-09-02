// Package auth verifies the same HS256 access tokens stakeholders issues,
// same pattern as tours/blog/followers/payments - gin-native here since
// encounters uses gin, not net/http+gorilla/mux like tours does.
package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Identity struct {
	Username string
	Role     string
}

type claims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type Verifier struct {
	secret []byte
}

func NewVerifier() (*Verifier, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return &Verifier{secret: []byte(secret)}, nil
}

func (v *Verifier) parse(tokenString string) (*Identity, error) {
	if tokenString == "" {
		return nil, errors.New("missing token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		return v.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := token.Claims.(*claims)
	if !ok || c.Subject == "" {
		return nil, errors.New("invalid claims")
	}
	return &Identity{Username: c.Subject, Role: c.Role}, nil
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// RequireAuth rejects the request with 401 unless it carries a valid token,
// then makes the identity available via CurrentIdentity(c).
func (v *Verifier) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := v.parse(extractBearerToken(c.GetHeader("Authorization")))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Set("username", id.Username)
		c.Set("role", id.Role)
		c.Next()
	}
}

// CurrentIdentity reads what RequireAuth set on the context.
func CurrentIdentity(c *gin.Context) Identity {
	return Identity{Username: c.GetString("username"), Role: c.GetString("role")}
}
