package services

import (
	"testing"
)

func TestHaversineDistance(t *testing.T) {
	// Gaziantep to Istanbul (rough distance ~850km)
	dist := haversineDistance(37.0902, 35.3872, 41.0082, 28.9784)

	// Actual haversine result for these coords is ~704km
	expected := 704000
	margin := 70400

	if dist < expected-margin || dist > expected+margin {
		t.Errorf("Expected distance around %d, got %d", expected, dist)
	}
}
