package restaurantrecommender

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRecommendHandler_NoQuery tests that the handler returns a 400 status
// when no "query" parameter is provided.
func TestRecommendHandler_NoQuery(t *testing.T) {
	// Create a dummy DB using sqlmock.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock DB: %v", err)
	}
	defer db.Close()

	// Expect call to retrieve restaurant styles and return an empty set.
	rows := sqlmock.NewRows([]string{"style"})
	mock.ExpectQuery("SELECT DISTINCT style FROM restaurants").
		WillReturnRows(rows)

	handler := RecommendHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/recommend", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	// Optionally check that all expected queries were met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet SQL expectations: %v", err)
	}
}

// TestRecommendHandler_ValidQuery tests a valid flow in which the database returns
// expected restaurant styles and a restaurant record.
func TestRecommendHandler_ValidQuery(t *testing.T) {
	// Create a mock DB.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock DB: %v", err)
	}
	defer db.Close()

	// Expect the SQL query to retrieve distinct restaurant styles.
	stylesRows := sqlmock.NewRows([]string{"style"}).
		AddRow("Italian")
	mock.ExpectQuery("SELECT DISTINCT style FROM restaurants").
		WillReturnRows(stylesRows)

	// Expect the SQL query to retrieve restaurant records.
	restaurantRows := sqlmock.NewRows([]string{
		"name", "style", "address", "openHour", "closeHour", "vegetarian", "deliveries",
	}).
		AddRow("Test Restaurant", "Italian", "123 Main St", "09:00", "23:00", true, true)
	mock.ExpectQuery("SELECT name, style, address, openHour, closeHour, vegetarian, deliveries FROM restaurants").
		WillReturnRows(restaurantRows)

	// Create a valid GET request with query parameter.
	req := httptest.NewRequest(http.MethodGet, "/recommend?query=Italian", nil)
	rec := httptest.NewRecorder()

	handler := RecommendHandler(db)
	handler(rec, req)

	// Check that we received a successful response.
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", rec.Code)
	}

	// Unmarshal the response JSON into a Recommendation struct.
	var resp Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Errorf("Error unmarshalling response JSON: %v", err)
	}

	// Verify that the expected restaurant is returned.
	if resp.RestaurantRecommendation.Name != "Test Restaurant" {
		t.Errorf("Expected restaurant 'Test Restaurant', got %s", resp.RestaurantRecommendation.Name)
	}

	// Ensure all expected SQL queries were executed.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet SQL expectations: %v", err)
	}
}
