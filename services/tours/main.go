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
	"tours/messaging"
	"tours/repository"
	"tours/rpc"
	"tours/service"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

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

	// Repositories
	tourRepo := repository.NewTourRepository(db)
	keyPointRepo := repository.NewKeyPointRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	touristPositionRepo := repository.NewTouristPositionRepository(db)
	execRepo := repository.NewTourExecutionRepository(db)
	purchaseRepo := repository.NewPurchaseRepository()

	// RabbitMQ consumers
    messaging.StartCheckoutConsumer(tourRepo)

    messaging.StartPurchaseCompletedConsumer(purchaseRepo)

	// Services
	tourService := service.NewTourService(tourRepo, keyPointRepo, purchaseRepo)
	keyPointService := service.NewKeyPointService(keyPointRepo, tourRepo)
	reviewService := service.NewReviewService(reviewRepo, tourRepo)
	touristPositionService := service.NewTouristPositionService(touristPositionRepo)
	execService := service.NewTourExecutionService(execRepo, tourRepo, keyPointRepo, purchaseRepo, touristPositionRepo,)

	// Handlers
	tourHandler := handler.NewTourHandler(tourService)
	keyPointHandler := handler.NewKeyPointHandler(keyPointService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	touristPositionHandler := handler.NewTouristPositionHandler(touristPositionService)
	execHandler := handler.NewTourExecutionHandler(execService)

	r := mux.NewRouter()

	// Tours
	r.HandleFunc("/tours", tourHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/tours", tourHandler.GetPublished).Methods(http.MethodGet)
	r.HandleFunc("/tours/author/{authorId}", tourHandler.GetByAuthor).Methods(http.MethodGet)
	r.HandleFunc("/tours/{id}", tourHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/tours/{id}", tourHandler.Update).Methods(http.MethodPut)
	r.HandleFunc("/tours/{id}/publish", tourHandler.Publish).Methods(http.MethodPut)
	r.HandleFunc("/tours/{id}/archive", tourHandler.Archive).Methods(http.MethodPut)
	r.HandleFunc("/tours/{id}/activate", tourHandler.Activate).Methods(http.MethodPut)

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

	// Tourist Position
	r.HandleFunc("/tourist-position", touristPositionHandler.Update).Methods(http.MethodPut)
	r.HandleFunc("/tourist-position/{touristId}", touristPositionHandler.GetByTouristID).Methods(http.MethodGet)

	// Tour Executions
	r.HandleFunc("/executions", execHandler.Start).Methods(http.MethodPost)
	r.HandleFunc("/executions/{id}", execHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/executions/tourist/{touristId}", execHandler.GetByTourist).Methods(http.MethodGet)
	r.HandleFunc("/executions/{id}/check-keypoint", execHandler.CheckKeyPoint).Methods(http.MethodPost)
	r.HandleFunc("/executions/{id}/complete", execHandler.Complete).Methods(http.MethodPut)
	r.HandleFunc("/executions/{id}/abandon", execHandler.Abandon).Methods(http.MethodPut)

	// Pokretanje gRPC servera u pozadini
	grpcPort := os.Getenv("TOURS_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9093"
	}
	go func() {
		if err := rpc.StartGRPCServer(grpcPort, tourService); err != nil {
			log.Fatalf("neuspesno pokretanje gRPC servera: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	fmt.Printf("Tour service running on port %s\n", port)
	fmt.Printf("Tour gRPC server running on port %s\n", grpcPort)
	go func() {
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down tour service...")
}
