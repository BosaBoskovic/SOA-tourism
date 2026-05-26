package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	stakeholdersv1 "soa-tourism-proto/stakeholders/v1"
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

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
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

func extractBearerToken(authHeader string) string {
	if strings.TrimSpace(authHeader) == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
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
	stakeholdersGRPCURL := getEnvOrDefault("STAKEHOLDERS_GRPC_URL", "localhost:9091")
	blogURL := getEnvOrDefault("BLOG_URL", "http://localhost:8082")
	followersURL := getEnvOrDefault("FOLLOWERS_URL", "http://localhost:8084")
	toursURL := getEnvOrDefault("TOURS_URL", "http://localhost:8085")

	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer grpcCancel()
	grpcConn, err := grpc.DialContext(
		grpcCtx,
		stakeholdersGRPCURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10*1024*1024)),
	)
	if err != nil {
		log.Fatalf("neuspesno povezivanje sa stakeholders gRPC servisom: %v", err)
	}
	defer grpcConn.Close()
	stakeholdersGrpcClient := stakeholdersv1.NewStakeholdersServiceClient(grpcConn)

	stakeholdersProxy := newReverseProxy(stakeholdersURL)
	blogProxy := newReverseProxy(blogURL)
	followersProxy := newReverseProxy(followersURL)
	toursProxy := newReverseProxy(toursURL)

	mux := http.NewServeMux()

	// --- Stakeholders servis ---
	mux.HandleFunc("/stakeholders/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Metod nije dozvoljen"})
			return
		}
		var reqPayload struct {
			UsernameOrEmail    string `json:"usernameOrEmail"`
			UsernameOrEmailAlt string `json:"username_or_email"`
			Password           string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Neispravan zahtev"})
			return
		}

		usernameOrEmail := strings.TrimSpace(reqPayload.UsernameOrEmail)
		if usernameOrEmail == "" {
			usernameOrEmail = strings.TrimSpace(reqPayload.UsernameOrEmailAlt)
		}

		req := stakeholdersv1.LoginRequest{
			UsernameOrEmail: usernameOrEmail,
			Password:        reqPayload.Password,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		resp, err := stakeholdersGrpcClient.Login(ctx, &req)
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri prijavi"})
				return
			}
			switch st.Code() {
			case codes.InvalidArgument:
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Neispravan zahtev"})
			case codes.Unauthenticated:
				if st.Message() == "invalid_credentials" {
					writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Pogresni kredencijali"})
					return
				}
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Neispravan zahtev"})
			case codes.PermissionDenied:
				writeJSON(w, http.StatusForbidden, map[string]any{"error": "Nalog je blokiran"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri prijavi"})
			}
			return
		}

		accountPayload := map[string]any{}
		if resp.Account != nil {
			accountPayload["username"] = resp.Account.Username
			accountPayload["email"] = resp.Account.Email
			accountPayload["role"] = resp.Account.Role
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"message":     "Uspesna prijava",
			"accessToken": resp.AccessToken,
			"tokenType":   resp.TokenType,
			"expiresIn":   resp.ExpiresIn,
			"expiresAt":   resp.ExpiresAt,
			"account":     accountPayload,
		})
	})
	mux.HandleFunc("/stakeholders/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Metod nije dozvoljen"})
			return
		}
		accessToken := extractBearerToken(r.Header.Get("Authorization"))
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		resp, err := stakeholdersGrpcClient.GetProfile(ctx, &stakeholdersv1.GetProfileRequest{
			AccessToken: accessToken,
		})
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Printf("[GATEWAY] GetProfile gRPC error: %v", err)
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri citanju profila"})
				return
			}
			log.Printf("[GATEWAY] GetProfile gRPC error: code=%s message=%s", st.Code(), st.Message())
			switch st.Code() {
			case codes.Unauthenticated:
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Neispravan ili istekao token"})
			case codes.NotFound:
				writeJSON(w, http.StatusNotFound, map[string]any{"error": "Profil nije pronadjen"})
			case codes.ResourceExhausted:
				writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "Profil je prevelik"})
			case codes.Unavailable:
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Servis profila nije dostupan"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri citanju profila"})
			}
			return
		}

		profilePayload := map[string]any{}
		if resp.Profile != nil {
			profilePayload["username"] = resp.Profile.Username
			profilePayload["firstName"] = resp.Profile.FirstName
			profilePayload["lastName"] = resp.Profile.LastName
			profilePayload["imageURL"] = resp.Profile.ImageUrl
			profilePayload["bio"] = resp.Profile.Bio
			profilePayload["motto"] = resp.Profile.Motto
		}

		writeJSON(w, http.StatusOK, map[string]any{"profile": profilePayload})
	})
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

	paymentsURL := getEnvOrDefault("PAYMENTS_URL", "http://localhost:8086")
	paymentsProxy := newReverseProxy(paymentsURL)

	mux.Handle("/shopping-cart", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> payments", r.Method, r.URL.Path)
		paymentsProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/shopping-cart/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> payments", r.Method, r.URL.Path)
		paymentsProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/checkout/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> payments", r.Method, r.URL.Path)
		paymentsProxy.ServeHTTP(w, r)
	})))

	mux.Handle("/executions", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/executions/", withUsername(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
