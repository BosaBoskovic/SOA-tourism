package rpc

import (
	"context"
	"log"
	"net"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stakeholdersv1 "soa-tourism-proto/stakeholders/v1"
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

// Login is intentionally not implemented here anymore (falls through to
// UnimplementedStakeholdersServiceServer, which returns a clean
// "unimplemented" gRPC error): the gateway now proxies /stakeholders/login
// straight to this service's own REST endpoint instead of calling this RPC,
// because the login response needs to carry a refreshToken and the
// LoginResponse proto message has no such field - regenerating it isn't
// possible without a protoc/buf toolchain in this environment.

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

	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	stakeholdersv1.RegisterStakeholdersServiceServer(server, NewStakeholdersGrpcServer(authSvc, profileSvc))

	log.Printf("Stakeholders gRPC server pokrenut na portu %s", port)
	return server.Serve(listener)
}
