package restaurantrecommender

import "time"

// Restaurant represents a restaurant record.
type Restaurant struct {
	Name       string `json:"name"`
	Style      string `json:"style"`
	Address    string `json:"address"`
	OpenHour   string `json:"openHour"`
	CloseHour  string `json:"closeHour"`
	Vegetarian bool   `json:"vegetarian"`
	Deliveries bool   `json:"deliveries"`
}

// Recommendation wraps the restaurant recommendation in a JSON object.
type Recommendation struct {
	RestaurantRecommendation Restaurant `json:"restaurantRecommendation"`
}

// QueryCriteria holds parsed filtering options from a natural language query.
type QueryCriteria struct {
	Style      string     // e.g., "Mexican", "Italian", etc.
	Vegetarian *bool      // nil if not specified.
	Delivers   *bool      // nil if not specified.
	OpenNow    bool       // true if "open now" is mentioned.
	OpenAt     *time.Time // specific time if provided (e.g., "open at 6pm")
}
