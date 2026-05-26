package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/yourname/location-assistant/go-service/internal/api"
	"github.com/yourname/location-assistant/go-service/internal/config"
	"github.com/yourname/location-assistant/go-service/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.New()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	router := api.SetupRouter(cfg, db)

	port := os.Getenv("GO_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
