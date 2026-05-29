package rpc

import (
	"context"
	"log"
	"net"
	"time"
	"tours/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	toursv1 "soa-tourism-proto/tours/v1"
)

// ToursGrpcServer implements the ToursService gRPC API.
type ToursGrpcServer struct {
	toursv1.UnimplementedToursServiceServer
	tourService *service.TourService
}

func NewToursGrpcServer(tourService *service.TourService) *ToursGrpcServer {
	return &ToursGrpcServer{tourService: tourService}
}

func (s *ToursGrpcServer) GetTour(ctx context.Context, req *toursv1.GetTourRequest) (*toursv1.GetTourResponse, error) {
	if req.TourId == "" {
		return nil, status.Error(codes.InvalidArgument, "tour_id is required")
	}

	tour, err := s.tourService.GetByID(req.TourId)
	if err != nil {
		if err.Error() == "tour not found" {
			return nil, status.Error(codes.NotFound, "tour_not_found")
		}
		if err.Error() == "invalid tour ID" {
			return nil, status.Error(codes.InvalidArgument, "invalid_tour_id")
		}
		return nil, status.Error(codes.Internal, "get_tour_failed")
	}

	pbDurations := make([]*toursv1.TourDuration, 0, len(tour.Durations))
	for _, d := range tour.Durations {
		pbDurations = append(pbDurations, &toursv1.TourDuration{
			Transport: string(d.Transport),
			Minutes:   int32(d.Minutes),
		})
	}

	publishedAt := ""
	if tour.PublishedAt != nil {
		publishedAt = tour.PublishedAt.Format(time.RFC3339)
	}

	return &toursv1.GetTourResponse{
		Tour: &toursv1.Tour{
			Id:          tour.ID.Hex(),
			AuthorId:    tour.AuthorID,
			Name:        tour.Name,
			Description: tour.Description,
			Difficulty:  string(tour.Difficulty),
			Tags:        tour.Tags,
			Status:      string(tour.Status),
			LengthKm:    tour.LengthKm,
			Durations:   pbDurations,
			Price:       tour.Price,
			CreatedAt:   tour.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   tour.UpdatedAt.Format(time.RFC3339),
			PublishedAt: publishedAt,
		},
	}, nil
}

func (s *ToursGrpcServer) GetPublishedTours(ctx context.Context, req *toursv1.GetPublishedToursRequest) (*toursv1.GetPublishedToursResponse, error) {
	previews, err := s.tourService.GetPublished()
	if err != nil {
		return nil, status.Error(codes.Internal, "get_published_tours_failed")
	}

	pbTours := make([]*toursv1.TourPreview, 0, len(previews))
	for _, p := range previews {
		publishedAt := ""
		if p.PublishedAt != nil {
			publishedAt = p.PublishedAt.Format(time.RFC3339)
		}

		preview := &toursv1.TourPreview{
			Id:          p.ID.Hex(),
			AuthorId:    p.AuthorID,
			Name:        p.Name,
			Description: p.Description,
			Difficulty:  string(p.Difficulty),
			Tags:        p.Tags,
			LengthKm:    p.LengthKm,
			Price:       p.Price,
			PublishedAt: publishedAt,
		}

		if p.FirstKeyPoint != nil {
			preview.FirstKeyPoint = &toursv1.KeyPointPreview{
				Id:        p.FirstKeyPoint.ID.Hex(),
				Name:      p.FirstKeyPoint.Name,
				Latitude:  p.FirstKeyPoint.Latitude,
				Longitude: p.FirstKeyPoint.Longitude,
			}
		}

		pbTours = append(pbTours, preview)
	}

	return &toursv1.GetPublishedToursResponse{Tours: pbTours}, nil
}

// StartGRPCServer starts the gRPC server on the specified port.
func StartGRPCServer(port string, tourService *service.TourService) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	toursv1.RegisterToursServiceServer(server, NewToursGrpcServer(tourService))

	log.Printf("Tours gRPC server pokrenut na portu %s", port)
	return server.Serve(listener)
}