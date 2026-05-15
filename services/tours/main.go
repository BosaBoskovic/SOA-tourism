package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tours/handler"
	"tours/repository"
	"tours/service"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// ── Database ──────────────────────────────────────────────
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://" + mongoURI + "/?connect=direct"))
	if err != nil {
		log.Fatal("MongoDB connect error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}
	fmt.Println("Connected to MongoDB!")

	db := client.Database("tourServiceDB")

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Println("MongoDB disconnect error:", err)
		}
	}()

	// ── Repositories ──────────────────────────────────────────
	tourRepo := repository.NewTourRepository(db)
	keyPointRepo := repository.NewKeyPointRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// ── Services ──────────────────────────────────────────────
	tourService := service.NewTourService(tourRepo)
	keyPointService := service.NewKeyPointService(keyPointRepo, tourRepo)
	reviewService := service.NewReviewService(reviewRepo, tourRepo)

	// ── Handlers ──────────────────────────────────────────────
	tourHandler := handler.NewTourHandler(tourService)
	keyPointHandler := handler.NewKeyPointHandler(keyPointService)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// ── Router ────────────────────────────────────────────────
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	// Tours
	r.HandleFunc("/tours", tourHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/tours/author/{authorId}", tourHandler.GetByAuthor).Methods(http.MethodGet)
	r.HandleFunc("/tours/{id}", tourHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/tours/{id}", tourHandler.Update).Methods(http.MethodPut)

	// KeyPoints
	r.HandleFunc("/keypoints", keyPointHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/keypoints/tour/{tourId}", keyPointHandler.GetByTour).Methods(http.MethodGet)
	r.HandleFunc("/keypoints/{id}", keyPointHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/keypoints/{id}", keyPointHandler.Update).Methods(http.MethodPut)
	r.HandleFunc("/keypoints/{id}", keyPointHandler.Delete).Methods(http.MethodDelete)

	// Reviews
	r.HandleFunc("/reviews", reviewHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/reviews/tour/{tourId}", reviewHandler.GetByTour).Methods(http.MethodGet)
	r.HandleFunc("/reviews/{id}", reviewHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/reviews/{id}", reviewHandler.Delete).Methods(http.MethodDelete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	fmt.Printf("Tour service running on port %s\n", port)

	go func() {
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down tour service...")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
