package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

type GeocodingHandler struct {
	geocodingService *services.GeocodingService
}

// NewGeocodingHandler creates a GeocodingHandler backed by the given GeocodingService.
// Deprecated: use LocationHandler instead, which handles both POST and GET reverse geocoding.
func NewGeocodingHandler(geocodingService *services.GeocodingService) *GeocodingHandler {
	return &GeocodingHandler{geocodingService: geocodingService}
}

// ReverseGeocode handles GET /geocode/reverse?lat=&lng=.
// Deprecated: use LocationHandler.ReverseGeocode, which is registered in the router.
func (h *GeocodingHandler) ReverseGeocode(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	result, err := h.geocodingService.ReverseGeocode(c.Request.Context(), lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
