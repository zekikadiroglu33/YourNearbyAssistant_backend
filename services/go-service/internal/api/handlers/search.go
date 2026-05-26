package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/models"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

type SearchHandler struct {
	aiClient      *services.AIClient
	placesService *services.PlacesService
}

func NewSearchHandler(aiClient *services.AIClient, placesService *services.PlacesService) *SearchHandler {
	return &SearchHandler{
		aiClient:      aiClient,
		placesService: placesService,
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Extract intent from user query via Python AI service
	// Trim history to last 6 messages (3 pairs) server-side
	history := req.ConversationHistory
	if len(history) > 6 {
		history = history[len(history)-6:]
	}

	intent, err := h.aiClient.ExtractIntent(services.IntentRequest{
		UserMessage:         req.Query,
		UserLatitude:        req.Latitude,
		UserLongitude:       req.Longitude,
		ConversationHistory: history,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "intent extraction failed: " + err.Error()})
		return
	}

	// 2. Build places search request from extracted intent
	placesReq := models.NearbySearchRequest{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Radius:    intent.RadiusMeters,
	}
	if intent.PlaceType != "" {
		placesReq.Type = &intent.PlaceType
	}
	if intent.Keyword != "" {
		placesReq.Keyword = &intent.Keyword
	}

	// 3. Search nearby places
	places, err := h.placesService.NearbySearch(placesReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "places search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Intent: models.SearchIntent{
			PlaceType:        intent.PlaceType,
			Keyword:          intent.Keyword,
			RadiusMeters:     intent.RadiusMeters,
			ResponseLanguage: intent.ResponseLanguage,
		},
		Places: places,
	})
}
