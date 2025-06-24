package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"geoip-service/internal/api"
	"geoip-service/internal/config"
	"geoip-service/internal/geoip"
	grpcServer "geoip-service/internal/grpc"
	pb "geoip-service/proto"

	"google.golang.org/grpc"
)

func main() {
	configuration := config.Load()

	logger := setupLogger(configuration.LogLevel)
	logger.Info("Starting GeoIP Service", "http_port", configuration.Port, "grpc_port", configuration.GRPCPort, "db_path", configuration.DBPath)

	geoipService, err := geoip.NewService(configuration.DBPath, logger)
	if err != nil {
		logger.Error("Failed to initialize GeoIP service", "error", err)
		os.Exit(1)
	}
	defer geoipService.Close()

	httpHandler := api.NewHandler(geoipService, logger)
	httpRouter := api.NewRouter(httpHandler, logger)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", configuration.Port),
		Handler:      httpRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", configuration.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen on gRPC port", "error", err, "port", configuration.GRPCPort)
		os.Exit(1)
	}

	grpcServerInstance := grpc.NewServer()
	grpcServiceHandler := grpcServer.NewServer(geoipService, logger)
	pb.RegisterGeoIPServiceServer(grpcServerInstance, grpcServiceHandler)

	var serverWaitGroup sync.WaitGroup

	serverWaitGroup.Add(1)
	go func() {
		defer serverWaitGroup.Done()
		logger.Info("HTTP server starting", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	serverWaitGroup.Add(1)
	go func() {
		defer serverWaitGroup.Done()
		logger.Info("gRPC server starting", "address", grpcListener.Addr().String())
		if err := grpcServerInstance.Serve(grpcListener); err != nil {
			logger.Error("gRPC server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	logger.Info("Shutting down servers...")

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		logger.Error("HTTP server forced to shutdown", "error", err)
	}

	grpcServerInstance.GracefulStop()

	logger.Info("Servers exited")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
