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

func NewGeocodingHandler(geocodingService *services.GeocodingService) *GeocodingHandler {
	return &GeocodingHandler{geocodingService: geocodingService}
}

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

	result, err := h.geocodingService.ReverseGeocode(lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
