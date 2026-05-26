package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/models"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

type PlacesHandler struct {
	placesService *services.PlacesService
	db            *sql.DB
}

func NewPlacesHandler(placesService *services.PlacesService, db *sql.DB) *PlacesHandler {
	return &PlacesHandler{
		placesService: placesService,
		db:            db,
	}
}

func (h *PlacesHandler) NearbySearch(c *gin.Context) {
	var req models.NearbySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	places, err := h.placesService.NearbySearch(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, places)
}
