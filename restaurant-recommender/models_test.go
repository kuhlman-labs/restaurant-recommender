package restaurantrecommender

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRestaurantJSONMarshalling(t *testing.T) {
	restaurant := Restaurant{
		Name:       "Sample Restaurant",
		Style:      "Mexican",
		Address:    "456 Example Rd",
		OpenHour:   "10:00",
		CloseHour:  "22:00",
		Vegetarian: false,
		Deliveries: true,
	}

	data, err := json.Marshal(restaurant)
	if err != nil {
		t.Fatalf("Failed to marshal restaurant: %v", err)
	}

	jsonStr := string(data)
	expectedSubstring := `"name":"Sample Restaurant"`
	if !strings.Contains(jsonStr, expectedSubstring) {
		t.Errorf("Expected JSON to contain %q, got %s", expectedSubstring, jsonStr)
	}
}

func TestQueryCriteriaTimeField(t *testing.T) {
	now := time.Now()
	qc := QueryCriteria{
		OpenAt: &now,
	}

	if qc.OpenAt == nil {
		t.Error("Expected OpenAt to be set")
	}

	// Check that the stored time equals now (allowing a small delta).
	if now.Sub(*qc.OpenAt) > time.Millisecond {
		t.Errorf("Expected OpenAt time to be %v, got %v", now, *qc.OpenAt)
	}
}
