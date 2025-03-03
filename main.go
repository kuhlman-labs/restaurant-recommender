package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	restaurantrecommender "github.com/kuhlman-labs/restaurant-recommender/restaurant-recommender"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/azuread"
)

func main() {
	var db *sql.DB
	var err error

	// Retrieve environment variables for DB server and name.
	server := os.Getenv("DB_SERVER")
	database := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	if user != "" && password != "" {
		log.Fatal("Environment variables DB_USER and DB_PASSWORD should not be set when using managed identity.")
	}

	if server == "" || database == "" {
		log.Fatal("Missing environment variables DB_SERVER or DB_NAME")
	}

	//connString := fmt.Sprintf("Server=%s;Database=%s;", server, database)
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=1433;database=%s;fedauth=ActiveDirectoryServicePrincipal;", server, user, password, database)

	/*
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
	*/

	db, err = sql.Open(azuread.DriverName, connString)
	if err != nil {
		log.Fatalf("Error opening database connection with connection string %s: %v", connString, err)
	}

	defer db.Close()

	// Test the connection.
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database with connection string %s: %v", connString, err)
	}
	fmt.Println("Connected to Database...")

	/*
		//Seeding and creating tables now handled in the flyway migrations.

		// Create tables if they don't exist.
		if err := restaurantrecommender.CreateTables(db); err != nil {
			log.Fatal("Error creating tables:", err)
		}

		// Seed sample restaurant data.
		if err := restaurantrecommender.SeedRestaurants(db); err != nil {
			log.Fatal("Error seeding restaurant data:", err)
		}
	*/

	http.HandleFunc("/recommend", restaurantrecommender.RecommendHandler(db))
	fmt.Println("Restaurant recommendation service is running on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
