package models

type SearchRequest struct {
	Latitude            float64             `json:"latitude" binding:"required"`
	Longitude           float64             `json:"longitude" binding:"required"`
	Query               string              `json:"query" binding:"required"`
	ConversationHistory []map[string]string `json:"conversation_history"`
}

type SearchIntent struct {
	PlaceType        string `json:"place_type"`
	Keyword          string `json:"keyword"`
	RadiusMeters     int    `json:"radius_meters"`
	ResponseLanguage string `json:"response_language"`
}

type SearchResponse struct {
	Intent SearchIntent `json:"intent"`
	Places []Place      `json:"places"`
}
