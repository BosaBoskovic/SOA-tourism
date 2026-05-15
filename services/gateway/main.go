package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

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
	// Registracija, login, profili, nalozi
	mux.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/profiles/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[GATEWAY] %s %s -> stakeholders", r.Method, r.URL.Path)
		stakeholdersProxy.ServeHTTP(w, r)
	})

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
	mux.HandleFunc("/tours/", func(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("  /auth/*, /accounts/*, /profiles/* -> %s", stakeholdersURL)
	log.Printf("  /blog/*                            -> %s", blogURL)
	log.Printf("  /followers/*                       -> %s", followersURL)
	log.Printf("  /tours/*                           -> %s", toursURL)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Gateway nije mogao da se pokrene: %v", err)
	}
}
