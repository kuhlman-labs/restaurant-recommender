package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	restaurantrecommender "github.com/kuhlman-labs/restaurant-recommender/restaurant-recommender"
	_ "github.com/microsoft/go-mssqldb"
	mssql "github.com/microsoft/go-mssqldb"
)

func main() {
	var err error
	var db *sql.DB

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
	if err := restaurantrecommender.CreateTables(db); err != nil {
		log.Fatal("Error creating tables:", err)
	}

	// Seed sample restaurant data.
	if err := restaurantrecommender.SeedRestaurants(db); err != nil {
		log.Fatal("Error seeding restaurant data:", err)
	}

	http.HandleFunc("/recommend", restaurantrecommender.RecommendHandler(db))
	fmt.Println("Restaurant recommendation service is running on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
