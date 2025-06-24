package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"geoip-service/internal/geoip"
)

func TestHealthCheck(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := &Handler{logger: logger}

	request, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.HealthCheck(responseRecorder, request)

	if status := responseRecorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status to be 'healthy', got %v", response["status"])
	}
}

func TestCheckCountryBadRequest(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := &Handler{logger: logger}

	// Test with invalid JSON
	request, err := http.NewRequest("POST", "/v1/check", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.CheckCountry(responseRecorder, request)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCheckCountryMissingIP(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := &Handler{logger: logger}

	requestBody := geoip.CheckRequest{
		AllowedCountries: []string{"US"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	request, err := http.NewRequest("POST", "/v1/check", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.CheckCountry(responseRecorder, request)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response ErrorResponse
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	if response.Error != "missing_ip" {
		t.Errorf("expected error to be 'missing_ip', got %v", response.Error)
	}
}
