package models

type Place struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Vicinity         string   `json:"vicinity"`
	Rating           *float64 `json:"rating"`
	UserRatingsTotal int      `json:"user_ratings_total"`
	IsOpenNow        *bool    `json:"is_open_now"`
	DistanceMeters   int      `json:"distance_meters"`
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	PriceLevel       *int     `json:"price_level"`
	Types            []string `json:"types"`
	PhotoReference   *string  `json:"photo_reference"`
}

type NearbySearchRequest struct {
	Latitude  float64  `json:"latitude" binding:"required"`
	Longitude float64  `json:"longitude" binding:"required"`
	Radius    int      `json:"radius" binding:"required,min=1,max=50000"`
	Type      *string  `json:"type"`
	Keyword   *string  `json:"keyword"`
}
