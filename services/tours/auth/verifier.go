// Package auth verifies the same HS256 access tokens stakeholders issues,
// so tours no longer has to blindly trust an X-Username header set by the
// gateway (which itself never checks the JWT signature).
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const identityContextKey contextKey = "identity"

// Identity is the verified caller, taken from a valid access token's claims.
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

// NewVerifier reads JWT_SECRET from the environment. It must match the
// secret stakeholders signs tokens with.
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

// Middleware verifies the Authorization header, if present, and stores the
// resulting identity on the request context. It never rejects a request by
// itself - routes that need auth call RequireIdentity, so public GET routes
// keep working for anonymous callers.
func (v *Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, err := v.parse(extractBearerToken(r.Header.Get("Authorization"))); err == nil {
			r = r.WithContext(context.WithValue(r.Context(), identityContextKey, id))
		}
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// FromContext returns the verified caller identity, if the request carried
// a valid access token.
func FromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(identityContextKey).(*Identity)
	return id, ok
}
