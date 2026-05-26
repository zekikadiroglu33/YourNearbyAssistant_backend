package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/models"
)

type UsersHandler struct {
	db *sql.DB
}

// NewUsersHandler creates a UsersHandler backed by the given database connection.
func NewUsersHandler(db *sql.DB) *UsersHandler {
	return &UsersHandler{db: db}
}

// CreateUser handles POST /api/v1/users.
// Inserts a new user by email; if the email already exists, returns the existing record.
func (h *UsersHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.QueryRow(
		`INSERT INTO users (email) VALUES ($1)
		 ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		 RETURNING id, email, created_at`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/:id.
// Returns the user record for the given UUID; responds 404 if not found.
func (h *UsersHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	err := h.db.QueryRow(
		`SELECT id, email, created_at FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SavePlace handles POST /api/v1/users/:id/saved-places.
// Saves a place to the user's favourites list.
func (h *UsersHandler) SavePlace(c *gin.Context) {
	userID := c.Param("id")

	var req models.SavedPlace
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var saved models.SavedPlace
	err := h.db.QueryRow(
		`INSERT INTO saved_places (user_id, place_id, name, latitude, longitude)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, place_id, name, latitude, longitude, saved_at`,
		userID, req.PlaceID, req.Name, req.Latitude, req.Longitude,
	).Scan(&saved.ID, &saved.UserID, &saved.PlaceID, &saved.Name,
		&saved.Latitude, &saved.Longitude, &saved.SavedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, saved)
}

// GetSavedPlaces handles GET /api/v1/users/:id/saved-places.
// Returns all places the user has saved, ordered by most recently saved first.
func (h *UsersHandler) GetSavedPlaces(c *gin.Context) {
	userID := c.Param("id")

	rows, err := h.db.Query(
		`SELECT id, user_id, place_id, name, latitude, longitude, saved_at
		 FROM saved_places WHERE user_id = $1 ORDER BY saved_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	places := []models.SavedPlace{}
	for rows.Next() {
		var p models.SavedPlace
		if err := rows.Scan(&p.ID, &p.UserID, &p.PlaceID, &p.Name,
			&p.Latitude, &p.Longitude, &p.SavedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		places = append(places, p)
	}

	c.JSON(http.StatusOK, places)
}
