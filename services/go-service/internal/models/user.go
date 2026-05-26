package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type SavedPlace struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PlaceID   string    `json:"place_id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	SavedAt   time.Time `json:"saved_at"`
}
