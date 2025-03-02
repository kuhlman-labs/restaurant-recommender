package restaurantrecommender

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// RecommendHandler returns a handler that has access to the db dependency.
func RecommendHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		styles, err := getRestaurantStyles(db)
		if err != nil {
			http.Error(w, "Error retrieving restaurant styles", http.StatusInternalServerError)
			return
		}

		queryParam := r.URL.Query().Get("query")
		if queryParam == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}

		criteria := parseQuery(queryParam, styles)
		now := time.Now()
		restaurants, err := getRestaurants(db)
		if err != nil {
			http.Error(w, "Error retrieving restaurants", http.StatusInternalServerError)
			return
		}

		var candidate Restaurant
		found := false
		for _, rest := range restaurants {
			if restaurantMatchesCriteria(rest, criteria, now) {
				candidate = rest
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "No restaurant found matching the criteria", http.StatusNotFound)
			go logQueryAndResponse(queryParam, Recommendation{
				RestaurantRecommendation: Restaurant{
					Name:       "No match found",
					Style:      "",
					Address:    "",
					OpenHour:   "",
					CloseHour:  "",
					Vegetarian: false,
					Deliveries: false,
				},
			}, db)
			return
		}

		response := Recommendation{RestaurantRecommendation: candidate}
		// Log the query and response asynchronously.
		go logQueryAndResponse(queryParam, response, db)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
