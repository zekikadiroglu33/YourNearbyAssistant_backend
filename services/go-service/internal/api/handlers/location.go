package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/models"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

type LocationHandler struct {
	geocodingService *services.GeocodingService
}

// NewLocationHandler creates a LocationHandler backed by the given GeocodingService.
func NewLocationHandler(geocodingService *services.GeocodingService) *LocationHandler {
	return &LocationHandler{geocodingService: geocodingService}
}

// CurrentLocation handles POST /api/v1/location/current.
// Accepts a JSON body with latitude and longitude, and returns full location details
// (formatted address, neighborhood, district, city, country) via reverse geocoding.
func (h *LocationHandler) CurrentLocation(c *gin.Context) {
	var req models.LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := h.geocodingService.ReverseGeocode(req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ReverseGeocode handles GET /api/v1/geocode/reverse?lat=&lng=.
// Query-param style reverse geocode; returns the same LocationInfo as CurrentLocation.
func (h *LocationHandler) ReverseGeocode(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}
	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	info, err := h.geocodingService.ReverseGeocode(lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}
