package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnvOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func newReverseProxy(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Neispravna URL adresa za servis: %s, greška: %v", targetURL, err)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

func rewritePrefix(prefix, replacement string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = replacement + strings.TrimPrefix(r.URL.Path, prefix)
		next.ServeHTTP(w, r)
	}
}

// extractUsernameFromJWT čita Subject claim iz JWT tokena bez verifikacije potpisa.
// Verifikacija se radi u svakom servisu koji to treba — gateway samo prosljeđuje username.
func extractUsernameFromJWT(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	tokenString := strings.TrimSpace(parts[1])

	segments := strings.Split(tokenString, ".")
	if len(segments) != 3 {
		return ""
	}

	// Dodaj padding ako nedostaje
	payload := segments[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return ""
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return ""
	}

	// JWT standard: Subject je "sub" claim — stakeholders servis postavlja username kao Subject
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub
	}
	return ""
}

// withUsername je middleware koji iz JWT-a izvlači username i dodaje X-Username header
func withUsername(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := extractUsernameFromJWT(r.Header.Get("Authorization"))
		if username != "" {
			r.Header.Set("X-Username", username)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	stakeholdersURL := getEnvOrDefault("STAKEHOLDERS_URL", "http://localhost:8081")
	blogURL         := getEnvOrDefault("BLOG_URL", "http://localhost:8082")
	followersURL    := getEnvOrDefault("FOLLOWERS_URL", "http://localhost:8084")
	toursURL        := getEnvOrDefault("TOURS_URL", "http://localhost:8085")

	stakeholdersProxy := newReverseProxy(stakeholdersURL)
	blogProxy         := newReverseProxy(blogURL)
	followersProxy    := newReverseProxy(followersURL)
	toursProxy        := newReverseProxy(toursURL)

	mux := http.NewServeMux()

	// --- Stakeholders servis ---
	mux.Handle("/stakeholders", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/stakeholders/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})))

	mux.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders (legacy /auth)", r.Method, r.URL.Path)
		r.URL.Path = strings.Replace(r.URL.Path, "/auth/", "/stakeholders/", 1)
		stakeholdersProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/accounts/", rewritePrefix("/accounts/", "/stakeholders/accounts/", stakeholdersProxy))
	mux.HandleFunc("/profiles/", rewritePrefix("/profiles/", "/stakeholders/profile/", stakeholdersProxy))

	// --- Blog servis ---
	mux.Handle("/blog", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> blog", r.Method, r.URL.Path)
		blogProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/blog/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> blog", r.Method, r.URL.Path)
		blogProxy.ServeHTTP(w, r)
	})))

	// --- Followers servis ---
	mux.Handle("/followers/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> followers", r.Method, r.URL.Path)
		followersProxy.ServeHTTP(w, r)
	})))

	// --- Tours servis ---
	mux.Handle("/tours", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/tours/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/keypoints", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/keypoints/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/reviews", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/reviews/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/tourist-position", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/tourist-position/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))

	// Fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] Nepoznata ruta: %s %s", r.Method, r.URL.Path)
		http.Error(w, `{"error": "Ruta nije pronađena"}`, http.StatusNotFound)
	})

	port := getEnvOrDefault("PORT", "8080")
	log.Printf("API Gateway pokrenut na portu %s", port)
	log.Printf("  /stakeholders/* -> %s", stakeholdersURL)
	log.Printf("  /blog/*         -> %s", blogURL)
	log.Printf("  /followers/*    -> %s", followersURL)
	log.Printf("  /tours/*        -> %s", toursURL)

	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("Gateway nije mogao da se pokrene: %v", err)
	}
}