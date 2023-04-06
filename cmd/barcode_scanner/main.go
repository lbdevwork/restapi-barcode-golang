/*package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/api"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
)

var database *sql.DB

func main() {
	var err error

	dsn := os.Getenv("MYSQL_DSN")
	database, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v\n", err)
	}
	defer database.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/product/{barcode}", productHandler)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	barcode := chi.URLParam(r, "barcode")
	if barcode == "" || is12DigitNumber(barcode) {
		handleError(w, http.StatusBadRequest, "Invalid barcode", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	productInfo, err := db.FetchProduct(ctx, database, barcode)
	if err == nil {

		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Product not found in local database: %s\n", barcode)
			productInfo, err = api.FetchProduct(ctx, barcode)
			if err != nil {
				handleError(w, http.StatusInternalServerError, "Error fetching product from Open Food Facts API", err)
				return
			}
			err = db.StoreProduct(ctx, database, productInfo)
			if err != nil {
				handleError(w, http.StatusInternalServerError, "Error storing product in local database", err)
				return
			}
		} else {
			handleError(w, http.StatusInternalServerError, "Error fetching product from local database", err)
			return
		}
	}

	json.NewEncoder(w).Encode(productInfo)
}

func handleError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

func is12DigitNumber(value string) bool {
	// Create a regular expression to match 12-digit numbers
	regex, err := regexp.Compile(`^\d{12}$`)
	if err != nil {
		fmt.Printf("Error creating regular expression: %v\n", err)
		return false
	}

	// Check if the value matches the regular expression
	return regex.MatchString(value)
}
*/

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
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/api"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
)

var database *sql.DB

func main() {
	var err error

	// Create a connection to the database
	dsn := "root:@tcp(localhost:3306)/barcodes?parseTime=true" //os.Getenv("MYSQL_DSN")
	database, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v\n", err)
	}
	defer database.Close()

	// Create a router and register the middleware and routes
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/product/{barcode}", productHandler)
	})

	// Start the HTTP server on port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Console output of API status
	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// Handler for requests to the /v1/product/{barcode} endpoint
func productHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the barcode from the URL parameter and remove the .json suffix
	barcode := chi.URLParam(r, "barcode")
	barcode = strings.TrimSuffix(barcode, ".json")

	fmt.Printf("\n %v\n", barcode)

	// Check if the barcode is valid
	if barcode == "" || !is12DigitNumber(barcode) {
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
				handleError(w, http.StatusInternalServerError, "Error fetching product from Open Food Facts API", err)
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

// Helper function to check if a string is a 12-digit number
func is12DigitNumber(value string) bool {

	// Create a regular expression to match 12-digit numbers
	regex, err := regexp.Compile(`^\d{12}$`)
	if err != nil {
		fmt.Printf("Error creating regular expression: %v\n", err)
		return false
	}

	// Check if the value matches the regular expression
	return regex.MatchString(value)
}
