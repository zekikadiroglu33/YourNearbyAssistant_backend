package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AIClient struct {
	baseURL string
	client  *http.Client
}

// NewAIClient creates an AIClient that forwards requests to the Python AI service at baseURL.
func NewAIClient(baseURL string, client *http.Client) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		client:  client,
	}
}

type IntentRequest struct {
	UserMessage         string              `json:"user_message"`
	UserLatitude        float64             `json:"user_latitude"`
	UserLongitude       float64             `json:"user_longitude"`
	ConversationHistory []map[string]string `json:"conversation_history"`
}

type IntentResponse struct {
	PlaceType        string `json:"place_type"`
	Keyword          string `json:"keyword"`
	RadiusMeters     int    `json:"radius_meters"`
	ResponseLanguage string `json:"response_language"`
}

// ExtractIntent calls the Python AI service to parse a natural-language query into structured search intent.
func (c *AIClient) ExtractIntent(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/chat/intent", c.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read AI response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AI service error %d: %s", resp.StatusCode, string(respBody))
	}

	var intent IntentResponse
	if err := json.Unmarshal(respBody, &intent); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &intent, nil
}
