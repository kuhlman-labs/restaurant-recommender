package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	_ "github.com/microsoft/go-mssqldb"
	mssql "github.com/microsoft/go-mssqldb"
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

// Updated createTables creates the restaurants and query_logs tables for Azure SQL.
func createTables() error {
	restaurantTable := `
    IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'restaurants')
    BEGIN
      CREATE TABLE restaurants (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255) NOT NULL,
        style NVARCHAR(255) NOT NULL,
        address NVARCHAR(255),
        openHour NVARCHAR(5) NOT NULL,
        closeHour NVARCHAR(5) NOT NULL,
        vegetarian BIT,
        deliveries BIT
      );
    END;`
	_, err := db.Exec(restaurantTable)
	if err != nil {
		return err
	}

	queryLogsTable := `
    IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'query_logs')
    BEGIN
      CREATE TABLE query_logs (
        id INT IDENTITY(1,1) PRIMARY KEY,
        query NVARCHAR(MAX) NOT NULL,
        response NVARCHAR(MAX) NOT NULL,
        created_at DATETIME DEFAULT GETDATE()
      );
    END;`
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
		_, err := db.Exec(
			`INSERT INTO restaurants (name, style, address, openHour, closeHour, vegetarian, deliveries)
            VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`,
			sql.Named("p1", r.Name),
			sql.Named("p2", r.Style),
			sql.Named("p3", r.Address),
			sql.Named("p4", r.OpenHour),
			sql.Named("p5", r.CloseHour),
			sql.Named("p6", r.Vegetarian),
			sql.Named("p7", r.Deliveries),
		)
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

	_, err = db.Exec(
		"INSERT INTO query_logs (query, response, created_at) VALUES (@p1, @p2, @p3)",
		sql.Named("p1", query),
		sql.Named("p2", string(responseJSON)),
		sql.Named("p3", time.Now()),
	)
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
		})
		return
	}

	response := Recommendation{RestaurantRecommendation: candidate}

	// Log the query and response asynchronously.
	go logQueryAndResponse(queryParam, response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var db *sql.DB

func main() {
	var err error

	// Retrieve environment variables for DB server and name.
	server := os.Getenv("DB_SERVER")
	database := os.Getenv("DB_NAME")

	if server == "" || database == "" {
		log.Fatal("Missing environment variables DB_SERVER or DB_NAME")
	}

	connString := fmt.Sprintf("Server=%s;Database=%s", server, database)

	// Create a managed identity credential.
	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		log.Fatal("Error creating managed identity credential:", err.Error())
	}

	// Define a token provider function.
	tokenProvider := func() (string, error) {
		token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
			Scopes: []string{"https://database.windows.net//.default"},
		})
		return token.Token, err
	}

	// Create the Access Token Connector.
	connector, err := mssql.NewAccessTokenConnector(connString, tokenProvider)
	if err != nil {
		log.Fatal("Connector creation failed:", err.Error())
	}

	// Open the database connection.
	db = sql.OpenDB(connector)
	defer db.Close()

	// Test the connection.
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("Error pinging database:", err.Error())
	}

	fmt.Println("Connected to Azure SQL using Managed Identity!")

	// Create tables if they don't exist.
	if err := createTables(); err != nil {
		log.Fatal("Error creating tables:", err)
	}

	// Seed sample restaurant data.
	if err := seedRestaurants(); err != nil {
		log.Fatal("Error seeding restaurant data:", err)
	}

	http.HandleFunc("/recommend", recommendFromNaturalLanguage)
	fmt.Println("Restaurant recommendation service is running on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
