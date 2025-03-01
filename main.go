package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

var db *sql.DB

// createTables creates the restaurants and query_logs tables if they don't exist.
func createTables() error {
	restaurantTable := `
	CREATE TABLE IF NOT EXISTS restaurants (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		style TEXT NOT NULL,
		address TEXT,
		openHour TEXT NOT NULL,
		closeHour TEXT NOT NULL,
		vegetarian BOOLEAN,
		deliveries BOOLEAN
	);`
	_, err := db.Exec(restaurantTable)
	if err != nil {
		return err
	}

	queryLogsTable := `
	CREATE TABLE IF NOT EXISTS query_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		query TEXT NOT NULL,
		response TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(queryLogsTable)
	return err
}

// seedRestaurants inserts sample restaurant data if the restaurants table is empty.
func seedRestaurants() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	sampleRestaurants := []Restaurant{
		{
			Name:       "Pizza Hut",
			Style:      "Italian",
			Address:    "Wherever Street 99, Somewhere",
			OpenHour:   "09:00",
			CloseHour:  "23:00",
			Vegetarian: true,
			Deliveries: true,
		},
		{
			Name:       "Taco Bell",
			Style:      "Mexican",
			Address:    "123 Burrito Blvd, Somecity",
			OpenHour:   "10:00",
			CloseHour:  "22:00",
			Vegetarian: false,
			Deliveries: true,
		},
		{
			Name:       "Seoul Bites",
			Style:      "Korean",
			Address:    "123 Kimchi Ave, Seoul",
			OpenHour:   "11:00",
			CloseHour:  "22:00",
			Vegetarian: false,
			Deliveries: false,
		},
	}

	for _, r := range sampleRestaurants {
		_, err := db.Exec(`INSERT INTO restaurants (name, style, address, openHour, closeHour, vegetarian, deliveries)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			r.Name, r.Style, r.Address, r.OpenHour, r.CloseHour, r.Vegetarian, r.Deliveries)
		if err != nil {
			return err
		}
	}
	return nil
}

// getRestaurantStyles retrieves distinct restaurant styles from the database.
func getRestaurantStyles() ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT style FROM restaurants")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var styles []string
	for rows.Next() {
		var style string
		if err := rows.Scan(&style); err != nil {
			return nil, err
		}
		styles = append(styles, style)
	}
	return styles, nil
}

// getRestaurants retrieves all restaurant records.
func getRestaurants() ([]Restaurant, error) {
	rows, err := db.Query("SELECT name, style, address, openHour, closeHour, vegetarian, deliveries FROM restaurants")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []Restaurant
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(&r.Name, &r.Style, &r.Address, &r.OpenHour, &r.CloseHour, &r.Vegetarian, &r.Deliveries); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, r)
	}
	return restaurants, nil
}

// parseQuery extracts filtering criteria from a free‑form query using the provided styles.
func parseQuery(query string, styles []string) QueryCriteria {
	criteria := QueryCriteria{}
	lowerQuery := strings.ToLower(query)

	// Dynamically detect the restaurant style.
	for _, style := range styles {
		if strings.Contains(lowerQuery, strings.ToLower(style)) {
			criteria.Style = style
			break
		}
	}

	// Check for "vegetarian".
	if strings.Contains(lowerQuery, "vegetarian") {
		val := true
		criteria.Vegetarian = &val
	}

	// Check for delivery keywords.
	if strings.Contains(lowerQuery, "deliver") {
		val := true
		criteria.Delivers = &val
	}

	// Check for "open now".
	if strings.Contains(lowerQuery, "open now") {
		criteria.OpenNow = true
	}

	// Check for a specific time e.g. "open at 6pm" or "open at 6:30pm".
	re := regexp.MustCompile(`open at (\d{1,2})(?::(\d{2}))?\s*(am|pm)`)
	if matches := re.FindStringSubmatch(lowerQuery); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		minute := 0
		if matches[2] != "" {
			minute, _ = strconv.Atoi(matches[2])
		}
		ampm := matches[3]
		if ampm == "pm" && hour != 12 {
			hour += 12
		} else if ampm == "am" && hour == 12 {
			hour = 0
		}
		now := time.Now()
		openAt := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		criteria.OpenAt = &openAt
	}

	return criteria
}

// parseTime converts a string (e.g., "09:00") to a time.Time object.
func parseTime(tStr string) (time.Time, error) {
	return time.Parse("15:04", tStr)
}

// isOpen checks if a restaurant is open at the specified time.
func isOpen(r Restaurant, currentTime time.Time) bool {
	open, err := parseTime(r.OpenHour)
	if err != nil {
		return false
	}
	close, err := parseTime(r.CloseHour)
	if err != nil {
		return false
	}

	now := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, time.UTC)
	openTime := time.Date(0, 1, 1, open.Hour(), open.Minute(), 0, 0, time.UTC)
	closeTime := time.Date(0, 1, 1, close.Hour(), close.Minute(), 0, 0, time.UTC)

	// Adjust for restaurants that close past midnight.
	if closeTime.Before(openTime) {
		if now.Before(openTime) {
			now = now.Add(24 * time.Hour)
		}
		closeTime = closeTime.Add(24 * time.Hour)
	}
	return now.After(openTime) && now.Before(closeTime)
}

// restaurantMatchesCriteria returns true if a restaurant meets the query criteria.
func restaurantMatchesCriteria(r Restaurant, criteria QueryCriteria, now time.Time) bool {
	if criteria.Style != "" && strings.ToLower(r.Style) != strings.ToLower(criteria.Style) {
		return false
	}
	if criteria.Vegetarian != nil && r.Vegetarian != *criteria.Vegetarian {
		return false
	}
	if criteria.Delivers != nil && r.Deliveries != *criteria.Delivers {
		return false
	}
	var checkTime time.Time
	if criteria.OpenAt != nil {
		checkTime = *criteria.OpenAt
	} else if criteria.OpenNow {
		checkTime = now
	}
	if !checkTime.IsZero() && !isOpen(r, checkTime) {
		return false
	}
	return true
}

// logQueryAndResponse inserts the query and JSON response into the query_logs table.
func logQueryAndResponse(query string, response Recommendation) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	_, err = db.Exec("INSERT INTO query_logs (query, response, created_at) VALUES (?, ?, ?)",
		query, string(responseJSON), time.Now())
	if err != nil {
		log.Printf("Error logging query and response: %v", err)
	}
}

// recommendFromNaturalLanguage handles HTTP requests with a natural language query.
func recommendFromNaturalLanguage(w http.ResponseWriter, r *http.Request) {
	styles, err := getRestaurantStyles()
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
	restaurants, err := getRestaurants()
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
		return
	}

	response := Recommendation{RestaurantRecommendation: candidate}

	// Log the query and response asynchronously.
	go logQueryAndResponse(queryParam, response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	var err error
	// Open a connection to a local SQLite database file.
	db, err = sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal("Error connecting to SQLite database:", err)
	}
	defer db.Close()

	// Create tables if they don't exist.
	if err := createTables(); err != nil {
		log.Fatal("Error creating tables:", err)
	}

	// Seed sample restaurant data.
	if err := seedRestaurants(); err != nil {
		log.Fatal("Error seeding restaurant data:", err)
	}

	http.HandleFunc("/recommend", recommendFromNaturalLanguage)
	fmt.Println("Restaurant recommendation service (using SQLite) is running on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
