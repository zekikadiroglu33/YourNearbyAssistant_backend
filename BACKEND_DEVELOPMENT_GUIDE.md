# Backend Development Guide — AI Location Assistant (Microservices)

**Architecture:** Microservices  
**Services:** Go (Business Logic + Places API) + Python (AI with Gemini)  
**Database:** PostgreSQL  
**API Gateway:** Nginx or Traefik  
**Timeline:** 8–12 weeks

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Prerequisites](#prerequisites)
3. [Service 1: Go Business Logic Service](#service-1-go-business-logic-service)
4. [Service 2: Python AI Service](#service-2-python-ai-service)
5. [PostgreSQL Database](#postgresql-database)
6. [API Gateway Setup](#api-gateway-setup)
7. [Inter-Service Communication](#inter-service-communication)
8. [Deployment](#deployment)
9. [Testing](#testing)
10. [Monitoring & Observability](#monitoring--observability)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Mobile App (iOS)                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │  API Gateway   │ (Nginx/Traefik)
                    │  Port 80/443   │
                    └────────┬───────┘
                             │
                 ┌───────────┴──────────┐
                 │                      │
                 ▼                      ▼
    ┌────────────────────┐   ┌──────────────────────┐
    │  Go Service        │   │  Python AI Service   │
    │  Port 8080         │◄──┤  Port 8001           │
    │                    │   │                      │
    │  - Places API      │   │  - Gemini API        │
    │  - Business Logic  │   │  - Intent Extract    │
    │  - Geocoding       │   │  - Chat Streaming    │
    │  - User Management │   │                      │
    └────────┬───────────┘   └──────────────────────┘
             │
             ▼
    ┌────────────────────┐
    │   PostgreSQL       │
    │   Port 5432        │
    │                    │
    │  - users           │
    │  - saved_places    │
    │  - chat_history    │
    │  - api_usage       │
    └────────────────────┘
```

### Service Responsibilities

| Service | Port | Language | Responsibilities |
|---------|------|----------|-----------------|
| **Go Service** | 8080 | Go 1.21+ | Google Places API, Geocoding, Business logic, PostgreSQL queries, User auth |
| **Python AI Service** | 8001 | Python 3.11+ | Gemini intent extraction, Streaming chat responses, Multi-turn conversation |
| **API Gateway** | 80/443 | Nginx | Routing, SSL/TLS, Rate limiting, CORS |
| **PostgreSQL** | 5432 | PostgreSQL 16 | Users, Saved places, Chat history, API usage logs |

---

## Prerequisites

### Required Tools
- **Go 1.21+** (https://go.dev/dl/)
- **Python 3.11+** (https://www.python.org/downloads/)
- **PostgreSQL 16** (https://www.postgresql.org/download/)
- **Docker & Docker Compose** (for containerization)
- **Git** (version control)

### Required Accounts
- Google Cloud Platform (Places API key)
- Google AI Studio (Gemini API key)

### Environment Variables

Create `.env` file in project root:

```bash
# Database
DATABASE_URL=postgresql://postgres:password@localhost:5432/location_assistant

# Go Service
GO_SERVICE_PORT=8080
GOOGLE_PLACES_API_KEY=your_places_key_here

# Python AI Service
PYTHON_SERVICE_PORT=8001
GEMINI_API_KEY=your_gemini_key_here

# API Gateway
GATEWAY_PORT=80
```

---

## Service 1: Go Business Logic Service

**Directory:** `services/go-service/`

### Project Structure

```
services/go-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── places.go
│   │   │   ├── geocoding.go
│   │   │   ├── users.go
│   │   │   └── health.go
│   │   ├── middleware/
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   └── ratelimit.go
│   │   └── router.go
│   ├── models/
│   │   ├── place.go
│   │   ├── user.go
│   │   └── saved_place.go
│   ├── database/
│   │   ├── db.go
│   │   └── migrations/
│   │       ├── 001_create_users.sql
│   │       ├── 002_create_saved_places.sql
│   │       └── 003_create_chat_history.sql
│   ├── services/
│   │   ├── places_service.go
│   │   ├── geocoding_service.go
│   │   └── ai_client.go (HTTP client to Python service)
│   └── config/
│       └── config.go
├── pkg/
│   └── utils/
│       ├── distance.go
│       └── validation.go
├── go.mod
├── go.sum
├── Dockerfile
└── .env
```

### Step 1: Initialize Go Module

```bash
cd services/go-service
go mod init github.com/yourname/location-assistant/go-service
```

### Step 2: Install Dependencies

```bash
go get github.com/gin-gonic/gin
go get github.com/lib/pq
go get github.com/joho/godotenv
go get golang.org/x/time/rate
go get github.com/google/uuid
```

### Step 3: Main Entry Point

**File:** `cmd/server/main.go`

```go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourname/location-assistant/go-service/internal/api"
	"github.com/yourname/location-assistant/go-service/internal/config"
	"github.com/yourname/location-assistant/go-service/internal/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize config
	cfg := config.New()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	router := api.SetupRouter(cfg, db)

	// Start server
	port := os.Getenv("GO_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

### Step 4: Configuration

**File:** `internal/config/config.go`

```go
package config

import "os"

type Config struct {
	DatabaseURL       string
	GooglePlacesKey   string
	PythonServiceURL  string
	Port              string
}

func New() *Config {
	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		GooglePlacesKey:  os.Getenv("GOOGLE_PLACES_API_KEY"),
		PythonServiceURL: getEnv("PYTHON_SERVICE_URL", "http://localhost:8001"),
		Port:             getEnv("GO_SERVICE_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

### Step 5: Database Connection

**File:** `internal/database/db.go`

```go
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS saved_places (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			place_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			latitude FLOAT NOT NULL,
			longitude FLOAT NOT NULL,
			saved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS api_usage (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id),
			endpoint VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
```

### Step 6: Models

**File:** `internal/models/place.go`

```go
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
```

**File:** `internal/models/user.go`

```go
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
```

### Step 7: Places Service (Google Places API)

**File:** `internal/services/places_service.go`

```go
package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"

	"github.com/yourname/location-assistant/go-service/internal/models"
)

type PlacesService struct {
	apiKey string
	client *http.Client
}

func NewPlacesService(apiKey string) *PlacesService {
	return &PlacesService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Haversine distance calculation
func haversineDistance(lat1, lng1, lat2, lng2 float64) int {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(earthRadius * c)
}

func (s *PlacesService) NearbySearch(req models.NearbySearchRequest) ([]models.Place, error) {
	// Build URL
	baseURL := "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
	params := url.Values{}
	params.Add("location", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude))
	params.Add("radius", fmt.Sprintf("%d", req.Radius))
	params.Add("key", s.apiKey)

	if req.Type != nil {
		params.Add("type", *req.Type)
	}
	if req.Keyword != nil {
		params.Add("keyword", *req.Keyword)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make request
	resp, err := s.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch places: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var result struct {
		Results []struct {
			PlaceID   string  `json:"place_id"`
			Name      string  `json:"name"`
			Vicinity  string  `json:"vicinity"`
			Rating    float64 `json:"rating"`
			UserRatingsTotal int `json:"user_ratings_total"`
			OpeningHours *struct {
				OpenNow bool `json:"open_now"`
			} `json:"opening_hours"`
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			PriceLevel *int     `json:"price_level"`
			Types      []string `json:"types"`
			Photos     []struct {
				Reference string `json:"photo_reference"`
			} `json:"photos"`
		} `json:"results"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "OK" && result.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("places API error: %s", result.Status)
	}

	// Convert to models
	places := make([]models.Place, 0, len(result.Results))
	for _, r := range result.Results {
		place := models.Place{
			ID:               r.PlaceID,
			Name:             r.Name,
			Vicinity:         r.Vicinity,
			UserRatingsTotal: r.UserRatingsTotal,
			Latitude:         r.Geometry.Location.Lat,
			Longitude:        r.Geometry.Location.Lng,
			PriceLevel:       r.PriceLevel,
			Types:            r.Types,
		}

		if r.Rating > 0 {
			place.Rating = &r.Rating
		}

		if r.OpeningHours != nil {
			place.IsOpenNow = &r.OpeningHours.OpenNow
		}

		if len(r.Photos) > 0 {
			place.PhotoReference = &r.Photos[0].Reference
		}

		// Calculate distance
		place.DistanceMeters = haversineDistance(
			req.Latitude, req.Longitude,
			r.Geometry.Location.Lat, r.Geometry.Location.Lng,
		)

		places = append(places, place)
	}

	// Sort by distance
	for i := 0; i < len(places)-1; i++ {
		for j := i + 1; j < len(places); j++ {
			if places[i].DistanceMeters > places[j].DistanceMeters {
				places[i], places[j] = places[j], places[i]
			}
		}
	}

	return places, nil
}
```

### Step 8: API Handlers

**File:** `internal/api/handlers/places.go`

```go
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
```

### Step 9: Router Setup

**File:** `internal/api/router.go`

```go
package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yourname/location-assistant/go-service/internal/api/handlers"
	"github.com/yourname/location-assistant/go-service/internal/api/middleware"
	"github.com/yourname/location-assistant/go-service/internal/config"
	"github.com/yourname/location-assistant/go-service/internal/services"
)

func SetupRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	// Services
	placesService := services.NewPlacesService(cfg.GooglePlacesKey)

	// Handlers
	placesHandler := handlers.NewPlacesHandler(placesService, db)
	healthHandler := handlers.NewHealthHandler(db)

	// Routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
		v1.POST("/places/nearby", placesHandler.NearbySearch)
	}

	return router
}
```

### Step 10: Dockerfile

**File:** `Dockerfile`

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
```

---

## Service 2: Python AI Service

**Directory:** `services/python-service/`

### Project Structure

```
services/python-service/
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── config.py
│   ├── models/
│   │   ├── __init__.py
│   │   ├── chat.py
│   │   └── intent.py
│   ├── services/
│   │   ├── __init__.py
│   │   ├── gemini_service.py
│   │   └── intent_service.py
│   └── api/
│       ├── __init__.py
│       ├── routes.py
│       └── middleware.py
├── requirements.txt
├── Dockerfile
└── .env
```

### Step 1: Requirements

**File:** `requirements.txt`

```
fastapi==0.104.1
uvicorn[standard]==0.24.0
google-generativeai==0.3.1
pydantic==2.5.0
python-dotenv==1.0.0
httpx==0.25.1
```

### Step 2: Configuration

**File:** `app/config.py`

```python
import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
    GEMINI_API_KEY: str = os.getenv("GEMINI_API_KEY", "")
    GO_SERVICE_URL: str = os.getenv("GO_SERVICE_URL", "http://localhost:8080")
    PORT: int = int(os.getenv("PYTHON_SERVICE_PORT", "8001"))
    
settings = Settings()
```

### Step 3: Models

**File:** `app/models/intent.py`

```python
from pydantic import BaseModel
from typing import Optional

class Intent(BaseModel):
    place_type: str  # restaurant, mosque, park, etc.
    keyword: Optional[str] = None  # kebap, lahmacun, etc.
    radius_meters: int = 1000
    response_language: str = "en"

class IntentRequest(BaseModel):
    user_message: str
    user_latitude: float
    user_longitude: float
    conversation_history: list[dict[str, str]] = []
```

**File:** `app/models/chat.py`

```python
from pydantic import BaseModel
from typing import List, Optional

class ChatMessage(BaseModel):
    role: str  # user or assistant
    content: str

class ChatRequest(BaseModel):
    message: str
    conversation_history: List[ChatMessage] = []
    user_latitude: float
    user_longitude: float

class ChatStreamChunk(BaseModel):
    text: str
```

### Step 4: Gemini Service

**File:** `app/services/gemini_service.py`

```python
import google.generativeai as genai
from app.config import settings
import json

class GeminiService:
    def __init__(self):
        genai.configure(api_key=settings.GEMINI_API_KEY)
        self.model = genai.GenerativeModel('gemini-1.5-flash')
    
    async def extract_intent(
        self,
        user_message: str,
        user_lat: float,
        user_lng: float,
        conversation_history: list
    ) -> dict:
        """Extract search intent from user message using Gemini."""
        
        system_prompt = f"""You are a location assistant. The user is at coordinates {user_lat:.5f}, {user_lng:.5f}.
Extract their search intent and respond ONLY with a JSON object — no markdown, no explanation:
{{
  "place_type": "<Google Places API type: restaurant|cafe|mosque|park|hospital|pharmacy|supermarket|bank|atm|bakery|bar|night_club|tourist_attraction|museum|school|gas_station>",
  "keyword": "<specific keyword to refine search, e.g. 'kebap', 'lahmacun', 'döner', or empty string>",
  "radius_meters": <number 300-5000>,
  "response_language": "<ISO 639-1 language code the user wrote in>"
}}

Rules:
- For kebap/döner/lahmacun/baklava etc → place_type: "restaurant", keyword: the food name
- For broad requests like "restaurants near me" → keyword: ""
- Larger radius if user says "around here", smaller if "very close" or "walking distance"
- Default radius: 1000
- Detect language from user input (en, tr, ar, etc.)"""

        # Build conversation context
        messages = []
        for msg in conversation_history[-6:]:  # Last 6 turns
            messages.append(f"{msg['role']}: {msg['content']}")
        messages.append(f"user: {user_message}")
        
        full_prompt = f"{system_prompt}\n\nConversation:\n" + "\n".join(messages)
        
        # Generate response
        response = self.model.generate_content(full_prompt)
        
        # Parse JSON response
        try:
            intent = json.loads(response.text.strip())
            return intent
        except json.JSONDecodeError:
            # Fallback if model doesn't return valid JSON
            return {
                "place_type": "restaurant",
                "keyword": "",
                "radius_meters": 1000,
                "response_language": "en"
            }
    
    async def stream_chat_response(
        self,
        user_message: str,
        places: list,
        intent: dict,
        user_lat: float,
        user_lng: float,
        conversation_history: list
    ):
        """Generate streaming chat response using Gemini."""
        
        # Build places context
        if not places:
            places_context = "No places were found nearby for this search."
        else:
            places_list = []
            for i, place in enumerate(places[:5], 1):  # Top 5 places
                distance = self._format_distance(place.get("distance_meters", 0))
                rating_text = ""
                if place.get("rating"):
                    rating_text = f", rated {place['rating']}/5"
                
                open_text = ""
                if place.get("is_open_now") is not None:
                    open_text = ", open now" if place["is_open_now"] else ", currently closed"
                
                places_list.append(
                    f"{i}. {place['name']} — {distance} away{rating_text}{open_text}\n"
                    f"   Address: {place['vicinity']}"
                )
            places_context = "\n".join(places_list)
        
        lang = intent.get("response_language", "en")
        
        system_prompt = f"""You are a friendly local guide assistant. The user is at {user_lat:.4f}, {user_lng:.4f}.
You have just searched nearby places and found the following results:

{places_context}

Respond in {lang} language. Be warm, concise, and helpful.
- Briefly introduce what you found (1 sentence)
- Mention the top 2-3 highlights naturally (closest, highest rated, open now)
- End with a helpful tip or question to refine further
- Do NOT list all results verbatim — the UI cards already show them
- Keep response under 120 words"""

        # Build messages
        messages = []
        for msg in conversation_history[-6:]:
            messages.append(f"{msg['role']}: {msg['content']}")
        messages.append(f"user: {user_message}")
        
        full_prompt = f"{system_prompt}\n\nConversation:\n" + "\n".join(messages)
        
        # Stream response
        response = self.model.generate_content(full_prompt, stream=True)
        
        for chunk in response:
            if chunk.text:
                yield chunk.text
    
    def _format_distance(self, meters: int) -> str:
        if meters >= 1000:
            return f"{meters / 1000:.1f} km"
        return f"{meters} m"

gemini_service = GeminiService()
```

### Step 5: API Routes

**File:** `app/api/routes.py`

```python
from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
from app.models.intent import IntentRequest
from app.models.chat import ChatRequest
from app.services.gemini_service import gemini_service
import httpx
from app.config import settings

router = APIRouter()

@router.post("/api/v1/chat/intent")
async def extract_intent(request: IntentRequest):
    """Extract search intent from user message."""
    try:
        intent = await gemini_service.extract_intent(
            user_message=request.user_message,
            user_lat=request.user_latitude,
            user_lng=request.user_longitude,
            conversation_history=request.conversation_history
        )
        return intent
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/api/v1/chat/stream")
async def stream_chat(request: ChatRequest):
    """Stream AI chat response with places context."""
    try:
        # 1. Extract intent
        intent = await gemini_service.extract_intent(
            user_message=request.message,
            user_lat=request.user_latitude,
            user_lng=request.user_longitude,
            conversation_history=[
                {"role": msg.role, "content": msg.content}
                for msg in request.conversation_history
            ]
        )
        
        # 2. Fetch places from Go service
        async with httpx.AsyncClient() as client:
            places_response = await client.post(
                f"{settings.GO_SERVICE_URL}/api/v1/places/nearby",
                json={
                    "latitude": request.user_latitude,
                    "longitude": request.user_longitude,
                    "radius": intent["radius_meters"],
                    "type": intent["place_type"] if intent["place_type"] else None,
                    "keyword": intent["keyword"] if intent["keyword"] else None
                }
            )
            places = places_response.json()
        
        # 3. Stream response
        async def generate():
            async for chunk in gemini_service.stream_chat_response(
                user_message=request.message,
                places=places,
                intent=intent,
                user_lat=request.user_latitude,
                user_lng=request.user_longitude,
                conversation_history=[
                    {"role": msg.role, "content": msg.content}
                    for msg in request.conversation_history
                ]
            ):
                yield f"data: {chunk}\n\n"
        
        return StreamingResponse(generate(), media_type="text/event-stream")
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health")
async def health_check():
    return {"status": "healthy", "service": "python-ai"}
```

### Step 6: Main Application

**File:** `app/main.py`

```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.routes import router
from app.config import settings
import uvicorn

app = FastAPI(title="AI Service", version="1.0.0")

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure properly in production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router)

if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=True
    )
```

### Step 7: Dockerfile

**File:** `Dockerfile`

```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8001

CMD ["python", "-m", "app.main"]
```

---

## PostgreSQL Database

### Schema Design

**File:** `database/schema.sql`

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Saved places table
CREATE TABLE saved_places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    place_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    saved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, place_id)
);

-- Chat history table
CREATE TABLE chat_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- API usage tracking
CREATE TABLE api_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    endpoint VARCHAR(255) NOT NULL,
    response_time_ms INT,
    status_code INT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_saved_places_user_id ON saved_places(user_id);
CREATE INDEX idx_chat_history_user_id ON chat_history(user_id);
CREATE INDEX idx_api_usage_user_id ON api_usage(user_id);
CREATE INDEX idx_api_usage_timestamp ON api_usage(timestamp);
```

---

## API Gateway Setup

### Nginx Configuration

**File:** `nginx/nginx.conf`

```nginx
upstream go_service {
    server go-service:8080;
}

upstream python_service {
    server python-service:8001;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
    limit_req zone=api_limit burst=20 nodelay;

    # Go service routes
    location /api/v1/places {
        proxy_pass http://go_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /api/v1/users {
        proxy_pass http://go_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Python AI service routes
    location /api/v1/chat {
        proxy_pass http://python_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Connection '';
        proxy_http_version 1.1;
        chunked_transfer_encoding off;
        proxy_buffering off;
        proxy_cache off;
    }

    # Health checks
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

---

## Docker Compose

**File:** `docker-compose.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: location_assistant
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  go-service:
    build:
      context: ./services/go-service
      dockerfile: Dockerfile
    environment:
      DATABASE_URL: postgresql://postgres:password@postgres:5432/location_assistant
      GOOGLE_PLACES_API_KEY: ${GOOGLE_PLACES_API_KEY}
      GO_SERVICE_PORT: 8080
      PYTHON_SERVICE_URL: http://python-service:8001
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy

  python-service:
    build:
      context: ./services/python-service
      dockerfile: Dockerfile
    environment:
      GEMINI_API_KEY: ${GEMINI_API_KEY}
      PYTHON_SERVICE_PORT: 8001
      GO_SERVICE_URL: http://go-service:8080
    ports:
      - "8001:8001"

  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/conf.d/default.conf
    ports:
      - "80:80"
    depends_on:
      - go-service
      - python-service

volumes:
  postgres_data:
```

---

## Deployment

### Production Deployment (Railway/Render)

1. **Push to GitHub:**
```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/yourusername/location-assistant-backend.git
git push -u origin main
```

2. **Deploy Go Service to Railway:**
   - Connect GitHub repo
   - Select `services/go-service`
   - Add environment variables
   - Deploy

3. **Deploy Python Service to Railway:**
   - Add new service
   - Select `services/python-service`
   - Add environment variables
   - Deploy

4. **Deploy PostgreSQL:**
   - Add PostgreSQL plugin in Railway
   - Copy connection string to Go service env

---

## Testing

### Go Service Tests

**File:** `services/go-service/internal/services/places_service_test.go`

```go
package services

import (
	"testing"
	"github.com/yourname/location-assistant/go-service/internal/models"
)

func TestHaversineDistance(t *testing.T) {
	// Gaziantep to Istanbul (rough distance ~850km)
	dist := haversineDistance(37.0902, 35.3872, 41.0082, 28.9784)
	
	// Allow 10% margin
	expected := 850000
	margin := 85000
	
	if dist < expected-margin || dist > expected+margin {
		t.Errorf("Expected distance around %d, got %d", expected, dist)
	}
}
```

### Python Service Tests

**File:** `services/python-service/tests/test_gemini.py`

```python
import pytest
from app.services.gemini_service import GeminiService

@pytest.mark.asyncio
async def test_extract_intent():
    service = GeminiService()
    
    intent = await service.extract_intent(
        user_message="Find me kebap restaurants",
        user_lat=37.0902,
        user_lng=35.3872,
        conversation_history=[]
    )
    
    assert intent["place_type"] == "restaurant"
    assert "kebap" in intent.get("keyword", "").lower()
```

---

## Monitoring & Observability

### Logging

Add structured logging to both services:

**Go:** Use `log/slog` (Go 1.21+)  
**Python:** Use `logging` module with JSON formatting

### Metrics

- Request count by endpoint
- Response times (p50, p95, p99)
- Error rates
- Database connection pool status

### Tools

- **Prometheus** for metrics
- **Grafana** for dashboards
- **Sentry** for error tracking

---

## Next Steps

1. **Week 1-2:** Setup Go service + PostgreSQL
2. **Week 3-4:** Implement Places API integration
3. **Week 5-6:** Setup Python AI service + Gemini
4. **Week 7-8:** Inter-service communication + testing
5. **Week 9-10:** API Gateway + deployment
6. **Week 11-12:** Monitoring + production hardening

**Resources:**
- [Go Documentation](https://go.dev/doc/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Google Gemini API](https://ai.google.dev/docs)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

**Questions?** Test each service independently before integration. Use Docker Compose for local development.
