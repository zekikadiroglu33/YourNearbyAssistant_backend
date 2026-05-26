package models

type LocationInfo struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	FormattedAddress string  `json:"formatted_address"`
	Neighborhood     string  `json:"neighborhood"`
	District         string  `json:"district"`
	City             string  `json:"city"`
	Country          string  `json:"country"`
	CountryCode      string  `json:"country_code"`
	PostalCode       string  `json:"postal_code"`
}

type LocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}
