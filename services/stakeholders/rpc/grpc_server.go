package rpc

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stakeholdersv1 "soa-tourism-proto/stakeholders/v1"
	"stakeholders/model"
	"stakeholders/service"
)

// StakeholdersGrpcServer implements the StakeholdersService gRPC API.
type StakeholdersGrpcServer struct {
	stakeholdersv1.UnimplementedStakeholdersServiceServer
	authSvc    *service.AuthService
	profileSvc *service.ProfileService
}

func NewStakeholdersGrpcServer(authSvc *service.AuthService, profileSvc *service.ProfileService) *StakeholdersGrpcServer {
	return &StakeholdersGrpcServer{authSvc: authSvc, profileSvc: profileSvc}
}

func (s *StakeholdersGrpcServer) Login(ctx context.Context, req *stakeholdersv1.LoginRequest) (*stakeholdersv1.LoginResponse, error) {
	usernameOrEmail := strings.TrimSpace(req.UsernameOrEmail)
	password := strings.TrimSpace(req.Password)
	if usernameOrEmail == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid_request")
	}

	token, expiresAt, acc, err := s.authSvc.Login(ctx, model.LoginRequest{
		UsernameOrEmail: usernameOrEmail,
		Password:        password,
	})
	if err != nil {
		switch err.Error() {
		case "invalid_credentials":
			return nil, status.Error(codes.Unauthenticated, "invalid_credentials")
		case "account_blocked":
			return nil, status.Error(codes.PermissionDenied, "account_blocked")
		default:
			return nil, status.Error(codes.Internal, "login_failed")
		}
	}

	return &stakeholdersv1.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64((15 * time.Minute).Seconds()),
		ExpiresAt:   expiresAt.Format(time.RFC3339),
		Account: &stakeholdersv1.AccountInfo{
			Username: acc.Username,
			Email:    acc.Email,
			Role:     acc.Role,
		},
	}, nil
}

func (s *StakeholdersGrpcServer) GetProfile(ctx context.Context, req *stakeholdersv1.GetProfileRequest) (*stakeholdersv1.GetProfileResponse, error) {
	token := strings.TrimSpace(req.AccessToken)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing_token")
	}

	claims, err := s.authSvc.ParseClaims(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid_token")
	}

	profile, err := s.profileSvc.GetProfile(ctx, claims.Subject)
	if err != nil {
		switch err.Error() {
		case "profile_not_found":
			return nil, status.Error(codes.NotFound, "profile_not_found")
		default:
			return nil, status.Error(codes.Internal, "profile_read_failed")
		}
	}

	return &stakeholdersv1.GetProfileResponse{
		Profile: &stakeholdersv1.Profile{
			Username:  profile.Username,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			ImageUrl:  profile.ImageURL,
			Bio:       profile.Bio,
			Motto:     profile.Motto,
		},
	}, nil
}

// StartGRPCServer starts the gRPC server on the specified port.
func StartGRPCServer(port string, authSvc *service.AuthService, profileSvc *service.ProfileService) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	stakeholdersv1.RegisterStakeholdersServiceServer(server, NewStakeholdersGrpcServer(authSvc, profileSvc))

	log.Printf("Stakeholders gRPC server pokrenut na portu %s", port)
	return server.Serve(listener)
}
