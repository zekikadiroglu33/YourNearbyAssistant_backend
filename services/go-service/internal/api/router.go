package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/api/handlers"
	"github.com/yourname/location-assistant/go-service/internal/api/middleware"
	"github.com/yourname/location-assistant/go-service/internal/config"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

// SetupRouter wires all services, handlers, and middleware into a configured Gin engine.
func SetupRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	placesService := services.NewPlacesService(cfg.GooglePlacesKey)
	geocodingService := services.NewGeocodingService(cfg.GooglePlacesKey)
	aiClient := services.NewAIClient(cfg.PythonServiceURL)

	placesHandler := handlers.NewPlacesHandler(placesService, db)
	healthHandler := handlers.NewHealthHandler(db)
	locationHandler := handlers.NewLocationHandler(geocodingService)
	usersHandler := handlers.NewUsersHandler(db)
	searchHandler := handlers.NewSearchHandler(aiClient, placesService)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)

		// Places
		v1.POST("/places/nearby", placesHandler.NearbySearch)

		// Smart search — location + natural language query → places
		v1.POST("/search", searchHandler.Search)

		// Location
		v1.GET("/location/current", locationHandler.CurrentLocation)
		v1.GET("/geocode/reverse", locationHandler.ReverseGeocode)

		// Users
		v1.POST("/users", usersHandler.CreateUser)
		v1.GET("/users/:id", usersHandler.GetUser)
		v1.POST("/users/:id/saved-places", usersHandler.SavePlace)
		v1.GET("/users/:id/saved-places", usersHandler.GetSavedPlaces)
	}

	return router
}
