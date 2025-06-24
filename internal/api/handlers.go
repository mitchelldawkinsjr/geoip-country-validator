package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"geoip-service/internal/geoip"
)

type Handler struct {
	geoipService *geoip.Service
	logger       *slog.Logger
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func NewHandler(geoipService *geoip.Service, logger *slog.Logger) *Handler {
	return &Handler{
		geoipService: geoipService,
		logger:       logger,
	}
}

func (handler *Handler) CheckCountry(writer http.ResponseWriter, request *http.Request) {
	var checkRequest geoip.CheckRequest

	if err := json.NewDecoder(request.Body).Decode(&checkRequest); err != nil {
		handler.logger.Error("Failed to decode request", "error", err)
		handler.writeErrorResponse(writer, http.StatusBadRequest, "invalid_request", "Invalid JSON format")
		return
	}

	if checkRequest.IPAddress == "" {
		handler.writeErrorResponse(writer, http.StatusBadRequest, "missing_ip", "IP address is required")
		return
	}

	if len(checkRequest.AllowedCountries) == 0 {
		handler.writeErrorResponse(writer, http.StatusBadRequest, "missing_countries", "At least one allowed country is required")
		return
	}

	result, err := handler.geoipService.CheckCountry(checkRequest)
	if err != nil {
		handler.logger.Error("Country check failed", "error", err, "ip", checkRequest.IPAddress)
		handler.writeErrorResponse(writer, http.StatusInternalServerError, "lookup_failed", "Failed to check country for IP address")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(result)
}

func (handler *Handler) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "geoip-service",
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(response)
}

func (handler *Handler) writeErrorResponse(writer http.ResponseWriter, statusCode int, errorCode, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	json.NewEncoder(writer).Encode(response)
}
