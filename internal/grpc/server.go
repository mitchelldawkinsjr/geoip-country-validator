package grpc

import (
	"context"
	"log/slog"

	"geoip-service/internal/geoip"
	pb "geoip-service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedGeoIPServiceServer
	geoipService *geoip.Service
	logger       *slog.Logger
}

func NewServer(geoipService *geoip.Service, logger *slog.Logger) *Server {
	return &Server{
		geoipService: geoipService,
		logger:       logger,
	}
}

func (server *Server) CheckCountry(context context.Context, grpcRequest *pb.CheckCountryRequest) (*pb.CheckCountryResponse, error) {
	if grpcRequest.IpAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "IP address is required")
	}

	if len(grpcRequest.AllowedCountries) == 0 {
		return nil, status.Error(codes.InvalidArgument, "At least one allowed country is required")
	}

	internalRequest := geoip.CheckRequest{
		IPAddress:        grpcRequest.IpAddress,
		AllowedCountries: grpcRequest.AllowedCountries,
	}

	checkResult, err := server.geoipService.CheckCountry(internalRequest)
	if err != nil {
		server.logger.Error("Country check failed", "error", err, "ip", grpcRequest.IpAddress)
		return nil, status.Error(codes.Internal, "Failed to check country for IP address")
	}

	grpcResponse := &pb.CheckCountryResponse{
		Allowed: checkResult.Allowed,
		Country: checkResult.Country,
	}

	return grpcResponse, nil
}

func (server *Server) Health(context context.Context, healthRequest *pb.HealthRequest) (*pb.HealthResponse, error) {
	healthResponse := &pb.HealthResponse{
		Status:  "healthy",
		Service: "geoip-service",
	}

	return healthResponse, nil
}
