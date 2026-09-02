// Package auth verifies the same HS256 access tokens stakeholders issues.
// The rest of the gateway deliberately does NOT verify JWTs itself (see
// extractUsernameFromJWT in main.go) - every proxied route relies on the
// backend service behind it to verify the Authorization header it forwards.
// This package exists only for the handful of endpoints the gateway itself
// serves directly (currently just the upload endpoint), which have nothing
// downstream to do that verification for them.
package auth

import (
	"errors"
	"os"
	"strings"

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

func (v *Verifier) Parse(authHeader string) (*Identity, error) {
	token := extractBearerToken(authHeader)
	if token == "" {
		return nil, errors.New("missing token")
	}
	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		return v.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := parsed.Claims.(*claims)
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
