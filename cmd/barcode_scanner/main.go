package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/api"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

var database *sql.DB

func main() {

	// Create a connection to the database
	database = connectToDatabase()
	defer database.Close()

	router := setupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Console output of API status
	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Handler for requests to the /v1/product/{barcode} endpoint
func productHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the barcode from the URL parameter and remove the .json suffix
	barcode := chi.URLParam(r, "barcode")
	barcode = strings.TrimSuffix(barcode, ".json")

	// Check if the barcode is valid
	if barcode == "" {
		handleError(w, http.StatusBadRequest, "Invalid barcode", nil)
		return
	}

	// Format to 13 digits
	barcode = utils.ConvertTo13DigitNumber(barcode)	

	if (barcode == "error"){	
		handleError(w, http.StatusBadRequest, "Invalid barcode", nil)
		return
	}

	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Fetch the product from the database
	product, err := db.FetchProduct(ctx, database, barcode)

	// If the product is not found, fetch it from the Open Food Facts API
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Product not found in local database: %s\n", barcode)

			// Fetch the product from the Open Food Facts API
			product, err = api.FetchProduct(ctx, barcode)
			if err != nil {
				handleError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching product from Open Food Facts API: %v", err), err)
				return
			}

			// If the product is not found, return a 404 Not Found error
			if product.ID == "" {
				handleError(w, http.StatusNotFound, "Barcode not found", nil)
				return
			}

			// Store the product in the database
			err = db.StoreProduct(ctx, database, product)
			if err != nil {
				handleError(w, http.StatusInternalServerError, "Error storing product in local database", err)
				return
			}

		} else {

			// If the error is not sql.ErrNoRows, return an internal server error ( Other error than Not Found )
			handleError(w, http.StatusInternalServerError, "Error fetching product from local database", err)
			return

		}
	} else {
		fmt.Printf("Product found in local database: %s\n", barcode)
	}

	json.NewEncoder(w).Encode(product)
}

// Handler function to return an error message as JSON
func handleError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}




func connectToDatabase() *sql.DB {
	dsn := "root:@tcp(localhost:3306)/barcodes?parseTime=true" //os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v\n", err)
	}
	return db
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/product/{barcode}", productHandler)
	})
	return r
}
