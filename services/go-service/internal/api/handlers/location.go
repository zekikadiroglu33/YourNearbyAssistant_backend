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

func NewLocationHandler(geocodingService *services.GeocodingService) *LocationHandler {
	return &LocationHandler{geocodingService: geocodingService}
}

// CurrentLocation resolves lat/lng sent by the client into full location details.
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

// ReverseGeocode handles GET /geocode/reverse?lat=&lng= (legacy query-param style).
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
