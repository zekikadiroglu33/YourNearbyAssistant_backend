# API Guide — YourNearbyAssistant Backend

**GitHub Repository:** https://github.com/zekikadiroglu33/YourNearbyAssistant_backend

**Base URL (Nginx Gateway):** `http://localhost`
**Go Service (direct):** `http://localhost:8080`
**Python AI Service (direct):** `http://localhost:8001`

All results verified and tested on 2026-05-26.

---

## App Scenario Flow

```
1. App starts          → Health Check
2. User opens app      → Register / Get User
3. Device gets GPS     → Detect Current Location
4. User types query    → Smart Search (AI + Places)
5. User refines query  → Smart Search with History
6. User likes a place  → Save Place
7. User views saves    → Get Saved Places
8. User wants AI chat  → Streaming Chat Response
```

---

## Step 1 — Health Check

Run this on app launch to confirm both services are up before making any calls.

### Go Service
```bash
curl -s http://localhost/api/v1/health
```
**Response:**
```json
{"service": "go-service", "status": "healthy"}
```

### Python AI Service
```bash
curl -s http://localhost:8001/health
```
**Response:**
```json
{"status": "healthy", "service": "python-ai"}
```

---

## Step 2 — Register / Get User

### Create User
The app calls this on first launch with the user's email. If the user already exists, it returns the existing record (upsert).

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```
**Response:**
```json
{
  "id": "09bb69ac-1e90-4f6d-9668-2972b5fd12b9",
  "email": "test@example.com",
  "created_at": "2026-05-26T19:14:03.358621Z"
}
```
> Save the `id`. It is required for saved places.

---

### Get User
Retrieve an existing user by ID (e.g. after app restart from stored UUID).

**Request:**
```bash
curl -s http://localhost/api/v1/users/09bb69ac-1e90-4f6d-9668-2972b5fd12b9
```
**Response:**
```json
{
  "id": "09bb69ac-1e90-4f6d-9668-2972b5fd12b9",
  "email": "test@example.com",
  "created_at": "2026-05-26T19:14:03.358621Z"
}
```

**Not found:**
```bash
curl -s http://localhost/api/v1/users/00000000-0000-0000-0000-000000000000
```
```json
{"error": "user not found"}
```

---

## Step 3 — Detect Current Location

The iOS app gets GPS coordinates from CoreLocation and passes them as query params. The backend resolves them into a human-readable location breakdown (address, district, city, country).

**Request:**
```bash
curl -s "http://localhost/api/v1/location/current?lat=41.0082&lng=28.9784"
```
**Response:**
```json
{
  "latitude": 41.0082,
  "longitude": 28.9784,
  "formatted_address": "Cankurtaran, Alemdar Cd., 34110 Fatih/İstanbul, Türkiye",
  "neighborhood": "",
  "district": "Fatih",
  "city": "İstanbul",
  "country": "Türkiye",
  "country_code": "TR",
  "postal_code": "34110"
}
```
> The app uses `city` and `district` to show the user "You are in Fatih, İstanbul" on the map header.

---

## Step 4 — Smart Search (First Query)

The main feature. User types anything in natural language — the backend uses Gemini to understand the intent, searches nearby places, and returns results. No need to pick a category or set a radius manually.

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "query": "I want kebap",
    "conversation_history": []
  }'
```
**Response:**
```json
{
  "intent": {
    "place_type": "restaurant",
    "keyword": "kebap",
    "radius_meters": 1000,
    "response_language": "en"
  },
  "places": [
    {
      "id": "ChIJg-twd4W5yhQR5sUBKnJ02Yw",
      "name": "KARADENİZ FİSH KEBAP PİDE HOUSE",
      "vicinity": "Alemdar, Dr. Eminpaşa Sk. No:3, 34110 Fatih/İstanbul, Türkiye",
      "rating": 4.9,
      "user_ratings_total": 1987,
      "is_open_now": true,
      "distance_meters": 264,
      "latitude": 41.0085416,
      "longitude": 28.9752846,
      "price_level": null,
      "types": ["turkish_restaurant", "family_restaurant", "restaurant", "food", "establishment"],
      "photo_reference": "places/ChIJg-twd4W5yhQR5sUBKnJ02Yw/photos/..."
    },
    {
      "id": "ChIJ5289_ZW5yhQRsQr7GgtCNNU",
      "name": "Saltanat Grill kebab & fish house",
      "vicinity": "Alemdar, Çatalçeşme Sk. NO:19/1A, 34110 Fatih/İstanbul, Türkiye",
      "rating": 4.9,
      "user_ratings_total": 4682,
      "is_open_now": true,
      "distance_meters": 264,
      "latitude": 41.0092137,
      "longitude": 28.9755442,
      "price_level": 2,
      "types": ["kebab_shop", "barbecue_restaurant", "restaurant", "food", "establishment"],
      "photo_reference": "places/ChIJ5289_ZW5yhQRsQr7GgtCNNU/photos/..."
    }
  ]
}
```
> `intent` shows what Gemini understood. `places` is sorted by distance (closest first). Up to 20 results.

---

## Step 5 — Smart Search with Follow-up (Conversation History)

The user refines their search in natural language. Pass the previous exchange in `conversation_history` — the backend trims it to the last 6 messages automatically before sending to Gemini.

### 5.1 — "Show me closer ones"
Gemini understands context, keeps `keyword: kebap` but shrinks the radius.

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "query": "show me closer ones",
    "conversation_history": [
      {"role": "user",      "content": "I want kebap"},
      {"role": "assistant", "content": "Found 20 kebap restaurants, closest is 264m away."}
    ]
  }'
```
**Response:**
```json
{
  "intent": {
    "place_type": "restaurant",
    "keyword": "kebap",
    "radius_meters": 300,
    "response_language": "en"
  },
  "places": [ "...20 places within 300m..." ]
}
```

---

### 5.2 — "What about cafes instead?"
Gemini switches the place type entirely based on context.

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "query": "what about cafes instead?",
    "conversation_history": [
      {"role": "user",      "content": "I want kebap"},
      {"role": "assistant", "content": "Found 20 kebap restaurants near you."}
    ]
  }'
```
**Response:**
```json
{
  "intent": {
    "place_type": "cafe",
    "keyword": "",
    "radius_meters": 1000,
    "response_language": "en"
  },
  "places": [
    {
      "name": "Hafız Mustafa",
      "distance_meters": 149,
      "rating": 4.9,
      "is_open_now": true
    },
    {
      "name": "Harab'be Cafe",
      "distance_meters": 191,
      "rating": 4.8,
      "is_open_now": true
    }
  ]
}
```

---

### 5.3 — Turkish follow-up
Language detection works on every turn independently.

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "query": "yakınımda kebapçı ara",
    "conversation_history": []
  }'
```
**Response:**
```json
{
  "intent": {
    "place_type": "restaurant",
    "keyword": "kebap",
    "radius_meters": 1000,
    "response_language": "tr"
  },
  "places": [ "...results..." ]
}
```

> **History trimming:** The client can send the full conversation — the backend always trims to the last 6 messages before passing to Gemini. No client-side management needed.

---

## Step 6 — Save a Place

After the user taps "Save" on a result card, call this endpoint with the place details.

**Request:**
```bash
curl -s -X POST http://localhost/api/v1/users/09bb69ac-1e90-4f6d-9668-2972b5fd12b9/saved-places \
  -H "Content-Type: application/json" \
  -d '{
    "place_id": "ChIJy5ROSeu5yhQRvNbPBu6f65U",
    "name": "360 Panorama Rooftop Restaurant",
    "latitude": 41.0090292,
    "longitude": 28.9774722
  }'
```
**Response:**
```json
{
  "id": "823a524f-e0a6-4c6a-8572-668a9c59aca8",
  "user_id": "09bb69ac-1e90-4f6d-9668-2972b5fd12b9",
  "place_id": "ChIJy5ROSeu5yhQRvNbPBu6f65U",
  "name": "360 Panorama Rooftop Restaurant",
  "latitude": 41.0090292,
  "longitude": 28.9774722,
  "saved_at": "2026-05-26T19:14:13.683639Z"
}
```

---

## Step 7 — Get Saved Places

Called when the user opens their bookmarks / saved places tab.

**Request:**
```bash
curl -s http://localhost/api/v1/users/09bb69ac-1e90-4f6d-9668-2972b5fd12b9/saved-places
```
**Response:**
```json
[
  {
    "id": "823a524f-e0a6-4c6a-8572-668a9c59aca8",
    "user_id": "09bb69ac-1e90-4f6d-9668-2972b5fd12b9",
    "place_id": "ChIJy5ROSeu5yhQRvNbPBu6f65U",
    "name": "360 Panorama Rooftop Restaurant",
    "latitude": 41.0090292,
    "longitude": 28.9774722,
    "saved_at": "2026-05-26T19:14:13.683639Z"
  }
]
```
> Results are ordered by `saved_at` descending (most recently saved first).

---

## Step 8 — Streaming AI Chat Response

For the full conversational experience — Gemini searches places and streams a natural language reply as SSE chunks. Used when the app shows a chat-style interface alongside the map.

### 8.1 — First message
```bash
curl -s -X POST http://localhost:8001/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "message": "find me a kebap restaurant nearby",
    "user_latitude": 41.0082,
    "user_longitude": 28.9784,
    "conversation_history": []
  }' --no-buffer
```
**Response (SSE stream):**
```
data: Great news! I've found a few excellent kebab restaurants very close to your location, all open and highly-rated at 4.9/5.

data:  The closest options are Saltanat Grill kebab & fish house and KARADENİZ FİSH KEBAP PİDE HOUSE, both just 264 meters away.

data:  Do any of these sound good, or would you like directions to one in particular?
```

### 8.2 — Follow-up in Turkish (language switches automatically)
```bash
curl -s -X POST http://localhost:8001/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "message": "peki kafe var mı?",
    "user_latitude": 41.0082,
    "user_longitude": 28.9784,
    "conversation_history": [
      {"role": "user",      "content": "find me a kebap restaurant nearby"},
      {"role": "assistant", "content": "Found several kebap restaurants near you..."}
    ]
  }' --no-buffer
```
**Response (SSE stream in Turkish):**
```
data: Merhaba! Yakınlarda birkaç harika kafe buldum, tam da kahve keyfi yapabileceğiniz yerler.

data:  En yakın ve şu anda açık olanlardan biri, sadece 149 metre mesafedeki 4.9 puanlı Hafız Mustafa.

data:  Nasıl bir kafe arıyordunuz, belki daha sakin bir ortam veya özel bir menüsü olan bir yer mi?
```

> **Stream internals:** intent extraction → Places API call → Gemini streaming response. Each `data:` chunk is appended to the chat bubble in real time.

---

## Advanced / Direct Endpoints

These are lower-level endpoints used by the app in specific situations.

---

### Direct Places Search

Use when you already know the exact type and radius (e.g. map filter UI) and don't need AI intent extraction.

**By type:**
```bash
curl -s -X POST http://localhost/api/v1/places/nearby \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "radius": 500,
    "type": "restaurant"
  }'
```

**By keyword + type:**
```bash
curl -s -X POST http://localhost/api/v1/places/nearby \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 41.0082,
    "longitude": 28.9784,
    "radius": 1000,
    "type": "restaurant",
    "keyword": "kebap"
  }'
```

**Response (first result):**
```json
[
  {
    "id": "ChIJy5ROSeu5yhQRvNbPBu6f65U",
    "name": "360 Panorama Rooftop Restaurant",
    "vicinity": "Alemdar, Yerebatan Cd. no:18, 34110 Fatih/İstanbul, Türkiye",
    "rating": 4.8,
    "user_ratings_total": 3869,
    "is_open_now": true,
    "distance_meters": 120,
    "latitude": 41.0090292,
    "longitude": 28.9774722,
    "price_level": null,
    "types": ["restaurant", "fine_dining_restaurant", "mediterranean_restaurant"],
    "photo_reference": null
  }
]
```

**Validation error** (`radius` must be 1–50000):
```bash
curl -s -X POST http://localhost/api/v1/places/nearby \
  -H "Content-Type: application/json" \
  -d '{"latitude": 41.0082, "longitude": 28.9784, "radius": 0}'
```
```json
{"error": "Key: 'NearbySearchRequest.Radius' Error:Field validation for 'Radius' failed on the 'required' tag"}
```

---

### AI Intent Only

Returns only what Gemini understood from the query — no places search. Useful for debugging or building custom search flows.

**English:**
```bash
curl -s -X POST http://localhost:8001/api/v1/chat/intent \
  -H "Content-Type: application/json" \
  -d '{
    "user_message": "find me a kebap restaurant nearby",
    "user_latitude": 41.0082,
    "user_longitude": 28.9784,
    "conversation_history": []
  }'
```
```json
{"place_type": "restaurant", "keyword": "kebap", "radius_meters": 1000, "response_language": "en"}
```

**Turkish:**
```bash
curl -s -X POST http://localhost:8001/api/v1/chat/intent \
  -H "Content-Type: application/json" \
  -d '{
    "user_message": "yakınımda kebapçı ara",
    "user_latitude": 41.0082,
    "user_longitude": 28.9784,
    "conversation_history": []
  }'
```
```json
{"place_type": "restaurant", "keyword": "kebap", "radius_meters": 1000, "response_language": "tr"}
```

---

### Reverse Geocode (query param style)

Alternative to `POST /api/v1/location/current` using query parameters.

```bash
curl -s "http://localhost/api/v1/geocode/reverse?lat=41.0082&lng=28.9784"
```
```json
{
  "latitude": 41.0082,
  "longitude": 28.9784,
  "formatted_address": "Cankurtaran, Alemdar Cd., 34110 Fatih/İstanbul, Türkiye",
  "neighborhood": "",
  "district": "Fatih",
  "city": "İstanbul",
  "country": "Türkiye",
  "country_code": "TR",
  "postal_code": "34110"
}
```

---

## Endpoint Reference

| # | Method | Endpoint | Service | Description | Status |
|---|--------|----------|---------|-------------|--------|
| 1 | GET | `/api/v1/health` | Go | Service health check | ✅ |
| 2 | GET | `/health` | Python :8001 | AI service health check | ✅ |
| 3 | POST | `/api/v1/users` | Go | Register / get user (upsert) | ✅ |
| 4 | GET | `/api/v1/users/:id` | Go | Get user by ID | ✅ |
| 5 | GET | `/api/v1/location/current?lat=&lng=` | Go | Detect current location from GPS | ✅ |
| 6 | POST | `/api/v1/search` | Go + Python | Smart search: query → intent → places | ✅ |
| 7 | POST | `/api/v1/users/:id/saved-places` | Go | Save a place for a user | ✅ |
| 8 | GET | `/api/v1/users/:id/saved-places` | Go | Get user's saved places | ✅ |
| 9 | POST | `/api/v1/chat/stream` | Python | Streaming AI chat response (SSE) | ✅ |
| 10 | POST | `/api/v1/places/nearby` | Go | Direct places search (no AI) | ✅ |
| 11 | POST | `/api/v1/chat/intent` | Python | AI intent extraction only | ✅ |
| 12 | GET | `/api/v1/geocode/reverse` | Go | Reverse geocode via query params | ✅ |
