package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yourname/location-assistant/go-service/internal/models"
)

type GeocodingService struct {
	apiKey string
	client *http.Client
}

// NewGeocodingService creates a GeocodingService authenticated with the given Google API key.
func NewGeocodingService(apiKey string, client *http.Client) *GeocodingService {
	return &GeocodingService{
		apiKey: apiKey,
		client: client,
	}
}

// ReverseGeocode converts GPS coordinates into a human-readable LocationInfo via the Google Geocoding API.
func (s *GeocodingService) ReverseGeocode(ctx context.Context, lat, lng float64) (*models.LocationInfo, error) {
	params := url.Values{}
	params.Add("latlng", fmt.Sprintf("%f,%f", lat, lng))
	params.Add("key", s.apiKey)

	fullURL := fmt.Sprintf("%s?%s", googleGeocodingURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create geocoding request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Results []struct {
			FormattedAddress  string `json:"formatted_address"`
			AddressComponents []struct {
				LongName  string   `json:"long_name"`
				ShortName string   `json:"short_name"`
				Types     []string `json:"types"`
			} `json:"address_components"`
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if result.Status != "OK" || len(result.Results) == 0 {
		return nil, fmt.Errorf("geocoding error: %s", result.Status)
	}

	r := result.Results[0]
	info := &models.LocationInfo{
		Latitude:         lat,
		Longitude:        lng,
		FormattedAddress: r.FormattedAddress,
	}

	// Parse address components
	for _, comp := range r.AddressComponents {
		types := comp.Types
		if contains(types, "neighborhood") || contains(types, "sublocality_level_1") {
			if info.Neighborhood == "" {
				info.Neighborhood = comp.LongName
			}
		}
		if contains(types, "administrative_area_level_2") {
			info.District = comp.LongName
		}
		if contains(types, "locality") {
			info.City = comp.LongName
		}
		if contains(types, "administrative_area_level_1") && info.City == "" {
			info.City = comp.LongName
		}
		if contains(types, "country") {
			info.Country = comp.LongName
			info.CountryCode = comp.ShortName
		}
		if contains(types, "postal_code") {
			info.PostalCode = comp.LongName
		}
	}

	return info, nil
}

// contains reports whether val is present in slice.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
