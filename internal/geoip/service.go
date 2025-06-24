package geoip

import (
	"fmt"
	"log/slog"
	"net"
	"slices"

	"github.com/oschwald/maxminddb-golang"
)

type Service struct {
	db     *maxminddb.Reader
	logger *slog.Logger
}

type CountryRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

type CheckRequest struct {
	IPAddress        string   `json:"ip_address"`
	AllowedCountries []string `json:"allowed_countries"`
}

type CheckResponse struct {
	Allowed bool   `json:"allowed"`
	Country string `json:"country"`
}

func NewService(dbPath string, logger *slog.Logger) (*Service, error) {
	database, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MaxMind database: %w", err)
	}

	logger.Info("GeoIP database loaded successfully", "path", dbPath)

	return &Service{
		db:     database,
		logger: logger,
	}, nil
}

func (service *Service) Close() error {
	return service.db.Close()
}

func (service *Service) CheckCountry(request CheckRequest) (*CheckResponse, error) {
	parsedIP := net.ParseIP(request.IPAddress)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", request.IPAddress)
	}

	var record CountryRecord
	err := service.db.Lookup(parsedIP, &record)
	if err != nil {
		service.logger.Error("Failed to lookup IP", "ip", request.IPAddress, "error", err)
		return nil, fmt.Errorf("failed to lookup IP: %w", err)
	}

	countryCode := record.Country.ISOCode

	if countryCode == "" {
		service.logger.Warn("No country found for IP", "ip", request.IPAddress)
		return &CheckResponse{
			Allowed: false,
			Country: "",
		}, nil
	}

	allowed := slices.Contains(request.AllowedCountries, countryCode)

	service.logger.Info("Country check performed",
		"ip", request.IPAddress,
		"country", countryCode,
		"allowed", allowed,
		"allowed_countries", request.AllowedCountries)

	return &CheckResponse{
		Allowed: allowed,
		Country: countryCode,
	}, nil
}
