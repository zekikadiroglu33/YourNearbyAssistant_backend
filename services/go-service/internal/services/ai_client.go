package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AIClient struct {
	baseURL string
	client  *http.Client
}

func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		client:  &http.Client{},
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

func (c *AIClient) ExtractIntent(req IntentRequest) (*IntentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		fmt.Sprintf("%s/api/v1/chat/intent", c.baseURL),
		"application/json",
		bytes.NewReader(body),
	)
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
