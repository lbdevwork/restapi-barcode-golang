package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

	// Checks if tables exist and creates them if they don't
	checkTables()

	// Define the routes for accessing the API
	router := setupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Console output of API status
	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fetchProduct(w, r, func(w http.ResponseWriter, product db.Product) {
		if product.ID == "" {
			handleError(w, http.StatusNotFound, "Barcode not found", nil)
			return
		}

		json.NewEncoder(w).Encode(product)
	})
}

func productHandlerText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fetchProduct(w, r, func(w http.ResponseWriter, product db.Product) {
		if product.ID == "" {
			handleError(w, http.StatusNotFound, "Barcode not found", nil)
			return
		}

		fmt.Fprintf(w, "Product ID: %s\nProduct Name: %s\nNutriscore Grade: %s\nEcoscore Grade: %s\n",
			product.ID, product.ProductName, product.NutriscoreGrade, product.EcoscoreGrade)
	})
}

// Handler function to return an error message as JSON
func handleError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

// Create a connection to the database
func connectToDatabase() *sql.DB {
	dsn := "root:@tcp(localhost:3306)/barcodes?parseTime=true" //os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v\n", err)
	}
	return db
}

// Define the routes for accessing the API
func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/product/{barcode}", productHandler)
		r.Get("/product/text/{barcode}.json", productHandlerText)
	})
	return r
}

// Checks if tables exist and creates them if they don't
func checkTables() {
	err := db.CreateSchema(database)
	if err != nil {
		log.Fatalf("Error creating database schema: %v\n", err)
	}
}

func fetchProduct(w http.ResponseWriter, r *http.Request, handleResponse func(http.ResponseWriter, db.Product)) {
	barcode := chi.URLParam(r, "barcode")
	barcode = strings.TrimSuffix(barcode, ".json")

	if barcode == "" {
		return
	}

	barcode = utils.ConvertTo13DigitNumber(barcode)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	product, err := db.FetchProduct(ctx, database, barcode)

	if err != nil {
		log.Printf("Product not found in local database: %s\n", barcode)

		product, err = api.FetchProduct(ctx, barcode)
		if err != nil {
			return
		}

		if product.ID == "" {
			return
		}

		err = db.StoreProduct(ctx, database, product)
		if err != nil {
			return
		}
	} else {
		fmt.Printf("Product found in local database: %s\n", barcode)
	}

	handleResponse(w, product)
}
