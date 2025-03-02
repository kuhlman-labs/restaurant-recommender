package restaurantrecommender

import (
	"testing"
	"time"
)

// boolPtr is a helper function to easily create *bool values.
func boolPtr(b bool) *bool {
	return &b
}

// TestParseQuery ensures that a free‑form query is parsed correctly.
func TestParseQuery(t *testing.T) {
	styles := []string{"Italian", "Mexican", "Korean"}
	query := "I am looking for an Italian restaurant that is vegetarian and open now"
	criteria := parseQuery(query, styles)

	if criteria.Style != "Italian" {
		t.Errorf("Expected style 'Italian', got %q", criteria.Style)
	}
	if criteria.Vegetarian == nil || !*criteria.Vegetarian {
		t.Error("Expected Vegetarian to be true")
	}
	if !criteria.OpenNow {
		t.Error("Expected OpenNow to be true")
	}
}

// TestParseTime checks that parsing the string time works.
func TestParseTime(t *testing.T) {
	tm, err := parseTime("09:00")
	if err != nil {
		t.Errorf("parseTime returned error: %v", err)
	}
	if tm.Hour() != 9 || tm.Minute() != 0 {
		t.Errorf("Expected time to be 9:00, got %v", tm.Format("15:04"))
	}
}

// TestRestaurantMatchesCriteria tests that the restaurant matching function
// returns true when criteria are met and false when they are not.
func TestRestaurantMatchesCriteria(t *testing.T) {
	// Create a sample restaurant.
	restaurant := Restaurant{
		Name:       "Test Restaurant",
		Style:      "Italian",
		Address:    "123 Main St",
		OpenHour:   "09:00",
		CloseHour:  "23:00",
		Vegetarian: true,
		Deliveries: true,
	}

	// Define criteria that should match.
	criteriaMatch := QueryCriteria{
		Style:      "Italian",
		Vegetarian: boolPtr(true),
		Delivers:   boolPtr(true),
		OpenNow:    true,
	}

	// Define a time during business hours.
	now := time.Date(2025, 3, 2, 12, 0, 0, 0, time.UTC)
	if !restaurantMatchesCriteria(restaurant, criteriaMatch, now) {
		t.Error("Expected restaurant to match criteria but it did not")
	}

	// Change criteria to one that should not match.
	criteriaNoMatch := QueryCriteria{
		Style:      "Mexican",
		Vegetarian: boolPtr(true),
		Delivers:   boolPtr(true),
		OpenNow:    true,
	}
	if restaurantMatchesCriteria(restaurant, criteriaNoMatch, now) {
		t.Error("Expected restaurant not to match criteria but it did")
	}
}
