package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// CORS middleware za omogućavanje cross-origin zahteva
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Odgovori na preflight zahteve
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

func main() {
	// URL-ovi servisa (čitaju se iz environment varijabli)
	stakeholdersURL := getEnvOrDefault("STAKEHOLDERS_URL", "http://localhost:8081")
	blogURL := getEnvOrDefault("BLOG_URL", "http://localhost:8082")
	followersURL := getEnvOrDefault("FOLLOWERS_URL", "http://localhost:8084")
	toursURL := getEnvOrDefault("TOURS_URL", "http://localhost:8085")

	// Kreiramo reverse proxy za svaki servis
	stakeholdersProxy := newReverseProxy(stakeholdersURL)
	blogProxy := newReverseProxy(blogURL)
	followersProxy := newReverseProxy(followersURL)
	toursProxy := newReverseProxy(toursURL)

	mux := http.NewServeMux()

	// --- Stakeholders servis ---
	// Front i backend koriste /stakeholders/* rute, pa gateway ne sme da menja prefiks.

	mux.HandleFunc("/stakeholders", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
    	stakeholdersProxy.ServeHTTP(w, r)
    })

	mux.HandleFunc("/stakeholders/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})

	// Kompatibilnost za starije klijente koji još šalju /auth/*, /accounts/* ili /profiles/*.
	mux.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders (legacy /auth)", r.Method, r.URL.Path)
		r.URL.Path = strings.Replace(r.URL.Path, "/auth/", "/stakeholders/", 1)
		stakeholdersProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/accounts/", rewritePrefix("/accounts/", "/stakeholders/accounts/", stakeholdersProxy))
	mux.HandleFunc("/profiles/", rewritePrefix("/profiles/", "/stakeholders/profile/", stakeholdersProxy))

	// --- Blog servis ---
	mux.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> blog", r.Method, r.URL.Path)
		blogProxy.ServeHTTP(w, r)
	})

	// --- Followers servis ---
	mux.HandleFunc("/followers/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> followers", r.Method, r.URL.Path)
		followersProxy.ServeHTTP(w, r)
	})

	// --- Tours servis ---

	mux.HandleFunc("/tours", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
    	toursProxy.ServeHTTP(w, r)
    })

	mux.HandleFunc("/tours/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
		toursProxy.ServeHTTP(w, r)
	})

    mux.HandleFunc("/keypoints", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
    	toursProxy.ServeHTTP(w, r)
    })

	mux.HandleFunc("/keypoints/", func(w http.ResponseWriter, r *http.Request) {
		toursProxy.ServeHTTP(w, r)
	})
    mux.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
    	toursProxy.ServeHTTP(w, r)
    })

	mux.HandleFunc("/reviews/", func(w http.ResponseWriter, r *http.Request) {
		toursProxy.ServeHTTP(w, r)
	})

    mux.HandleFunc("/tourist-position", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
    	toursProxy.ServeHTTP(w, r)
    })

    mux.HandleFunc("/tourist-position/", func(w http.ResponseWriter, r *http.Request) {
    	log.Printf("[GATEWAY] %s %s -> tours", r.Method, r.URL.Path)
    	toursProxy.ServeHTTP(w, r)
    })

	// Fallback - nepoznata ruta
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] Nepoznata ruta: %s %s", r.Method, r.URL.Path)
		http.Error(w, `{"error": "Ruta nije pronađena"}`, http.StatusNotFound)
	})

	port := getEnvOrDefault("PORT", "8080")
	log.Printf("API Gateway pokrenut na portu %s", port)
	log.Printf("  /stakeholders/*                  -> %s", stakeholdersURL)
	log.Printf("  /auth/*, /accounts/*, /profiles/* -> %s (legacy compat)", stakeholdersURL)
	log.Printf("  /blog/*                            -> %s", blogURL)
	log.Printf("  /followers/*                       -> %s", followersURL)
	log.Printf("  /tours/*                           -> %s", toursURL)

	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("Gateway nije mogao da se pokrene: %v", err)
	}
}
