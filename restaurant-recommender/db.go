package restaurantrecommender

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// Updated createTables creates the restaurants and query_logs tables for Azure SQL.
func CreateTables(db *sql.DB) error {
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
func SeedRestaurants(db *sql.DB) error {
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
func getRestaurantStyles(db *sql.DB) ([]string, error) {
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
func getRestaurants(db *sql.DB) ([]Restaurant, error) {
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

// logQueryAndResponse inserts the query and JSON response into the query_logs table.
func logQueryAndResponse(query string, response Recommendation, db *sql.DB) {
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
