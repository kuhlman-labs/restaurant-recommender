package restaurantrecommender

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestCreateTables tests that the CreateTables function executes the expected statements.
func TestCreateTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock DB: %v", err)
	}
	defer db.Close()

	// Expect the Exec for creating the restaurants table.
	mock.ExpectExec(regexp.QuoteMeta("IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'restaurants')")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Expect the Exec for creating the query_logs table.
	mock.ExpectExec(regexp.QuoteMeta("IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'query_logs')")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := CreateTables(db); err != nil {
		t.Errorf("CreateTables returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestSeedRestaurants_NoData tests that SeedRestaurants inserts data when the table is empty.
func TestSeedRestaurants_NoData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock DB: %v", err)
	}
	defer db.Close()

	// Expect a query counting rows.
	countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM restaurants")).
		WillReturnRows(countRows)

	// There are three sample restaurants so expect three INSERT statements.
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO restaurants (name, style, address, openHour, closeHour, vegetarian, deliveries)")).
		WithArgs("Pizza Hut", "Italian", "Wherever Street 99, Somewhere", "09:00", "23:00", true, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO restaurants (name, style, address, openHour, closeHour, vegetarian, deliveries)")).
		WithArgs("Taco Bell", "Mexican", "123 Burrito Blvd, Somecity", "10:00", "22:00", false, true).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO restaurants (name, style, address, openHour, closeHour, vegetarian, deliveries)")).
		WithArgs("Seoul Bites", "Korean", "123 Kimchi Ave, Seoul", "11:00", "22:00", false, false).
		WillReturnResult(sqlmock.NewResult(3, 1))

	if err := SeedRestaurants(db); err != nil {
		t.Errorf("SeedRestaurants returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestSeedRestaurants_WithData tests that when the table already has records,
// SeedRestaurants does nothing.
func TestSeedRestaurants_WithData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock DB: %v", err)
	}
	defer db.Close()

	// Return a non-zero count indicating data already exists.
	countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM restaurants")).
		WillReturnRows(countRows)

	if err := SeedRestaurants(db); err != nil {
		t.Errorf("SeedRestaurants returned error: %v", err)
	}

	// No INSERT expectations should be made.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestGetRestaurantStyles tests retrieving restaurant styles.
func TestGetRestaurantStyles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	styles := []string{"Italian", "Mexican", "Korean"}
	rows := sqlmock.NewRows([]string{"style"})
	for _, style := range styles {
		rows.AddRow(style)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT style FROM restaurants")).
		WillReturnRows(rows)

	gotStyles, err := getRestaurantStyles(db)
	if err != nil {
		t.Errorf("getRestaurantStyles returned error: %v", err)
	}
	if len(gotStyles) != len(styles) {
		t.Errorf("expected %d styles, got %d", len(styles), len(gotStyles))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestGetRestaurants tests retrieving restaurant records.
func TestGetRestaurants(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	restaurantRows := sqlmock.NewRows([]string{
		"name", "style", "address", "openHour", "closeHour", "vegetarian", "deliveries",
	}).
		AddRow("Pizza Hut", "Italian", "Wherever Street 99, Somewhere", "09:00", "23:00", true, true).
		AddRow("Taco Bell", "Mexican", "123 Burrito Blvd, Somecity", "10:00", "22:00", false, true)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT name, style, address, openHour, closeHour, vegetarian, deliveries FROM restaurants")).
		WillReturnRows(restaurantRows)

	restaurants, err := getRestaurants(db)
	if err != nil {
		t.Errorf("getRestaurants returned error: %v", err)
	}
	if len(restaurants) != 2 {
		t.Errorf("expected 2 restaurants, got %d", len(restaurants))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogQueryAndResponse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	resp := Recommendation{
		RestaurantRecommendation: Restaurant{
			Name:       "Pizza Hut",
			Style:      "Italian",
			Address:    "Wherever Street 99, Somewhere",
			OpenHour:   "09:00",
			CloseHour:  "23:00",
			Vegetarian: true,
			Deliveries: true,
		},
	}

	// Expect the Exec with the appropriate arguments.
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO query_logs (query, response, created_at) VALUES (@p1, @p2, @p3)")).
		WithArgs(
			"test query",
			sqlmock.AnyArg(), // response JSON (can't predict exact value)
			sqlmock.AnyArg(), // current time (any value)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	logQueryAndResponse("test query", resp, db)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
