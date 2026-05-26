package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/yourname/location-assistant/go-service/internal/models"
)

type PlacesService struct {
	apiKey string
	client *http.Client
}

func NewPlacesService(apiKey string) *PlacesService {
	return &PlacesService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) int {
	const earthRadius = 6371000

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(earthRadius * c)
}

// priceLevelToInt converts new API string price level to int.
func priceLevelToInt(s string) *int {
	m := map[string]int{
		"PRICE_LEVEL_FREE":           0,
		"PRICE_LEVEL_INEXPENSIVE":    1,
		"PRICE_LEVEL_MODERATE":       2,
		"PRICE_LEVEL_EXPENSIVE":      3,
		"PRICE_LEVEL_VERY_EXPENSIVE": 4,
	}
	if v, ok := m[s]; ok {
		return &v
	}
	return nil
}

func (s *PlacesService) NearbySearch(req models.NearbySearchRequest) ([]models.Place, error) {
	// Use searchText when keyword present, searchNearby otherwise
	if req.Keyword != nil && *req.Keyword != "" {
		return s.searchText(req)
	}
	return s.searchNearby(req)
}

// searchNearby uses Places API (New) — searchNearby endpoint.
func (s *PlacesService) searchNearby(req models.NearbySearchRequest) ([]models.Place, error) {
	body := map[string]any{
		"maxResultCount": 20,
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  req.Latitude,
					"longitude": req.Longitude,
				},
				"radius": float64(req.Radius),
			},
		},
	}
	if req.Type != nil && *req.Type != "" {
		body["includedTypes"] = []string{*req.Type}
	}

	return s.doRequest("https://places.googleapis.com/v1/places:searchNearby", body, req)
}

// searchText uses Places API (New) — searchText endpoint for keyword queries.
func (s *PlacesService) searchText(req models.NearbySearchRequest) ([]models.Place, error) {
	query := *req.Keyword
	if req.Type != nil && *req.Type != "" {
		query = query + " " + *req.Type
	}

	body := map[string]any{
		"textQuery":      query,
		"maxResultCount": 20,
		"locationBias": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  req.Latitude,
					"longitude": req.Longitude,
				},
				"radius": float64(req.Radius),
			},
		},
	}

	return s.doRequest("https://places.googleapis.com/v1/places:searchText", body, req)
}

func (s *PlacesService) doRequest(url string, body map[string]any, req models.NearbySearchRequest) ([]models.Place, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	fieldMask := "places.id,places.displayName,places.formattedAddress," +
		"places.rating,places.userRatingCount,places.currentOpeningHours," +
		"places.location,places.priceLevel,places.types,places.photos"

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Goog-Api-Key", s.apiKey)
	httpReq.Header.Set("X-Goog-FieldMask", fieldMask)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch places: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("places API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Places []struct {
			ID          string `json:"id"`
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			FormattedAddress      string  `json:"formattedAddress"`
			Rating                float64 `json:"rating"`
			UserRatingCount       int     `json:"userRatingCount"`
			CurrentOpeningHours   *struct {
				OpenNow bool `json:"openNow"`
			} `json:"currentOpeningHours"`
			Location struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			PriceLevel string   `json:"priceLevel"`
			Types      []string `json:"types"`
			Photos     []struct {
				Name string `json:"name"`
			} `json:"photos"`
		} `json:"places"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	places := make([]models.Place, 0, len(result.Places))
	for _, r := range result.Places {
		place := models.Place{
			ID:               r.ID,
			Name:             r.DisplayName.Text,
			Vicinity:         r.FormattedAddress,
			UserRatingsTotal: r.UserRatingCount,
			Latitude:         r.Location.Latitude,
			Longitude:        r.Location.Longitude,
			PriceLevel:       priceLevelToInt(r.PriceLevel),
			Types:            r.Types,
		}

		if r.Rating > 0 {
			place.Rating = &r.Rating
		}

		if r.CurrentOpeningHours != nil {
			place.IsOpenNow = &r.CurrentOpeningHours.OpenNow
		}

		if len(r.Photos) > 0 {
			place.PhotoReference = &r.Photos[0].Name
		}

		place.DistanceMeters = haversineDistance(
			req.Latitude, req.Longitude,
			r.Location.Latitude, r.Location.Longitude,
		)

		places = append(places, place)
	}

	// Sort by distance
	for i := 0; i < len(places)-1; i++ {
		for j := i + 1; j < len(places); j++ {
			if places[i].DistanceMeters > places[j].DistanceMeters {
				places[i], places[j] = places[j], places[i]
			}
		}
	}

	return places, nil
}
