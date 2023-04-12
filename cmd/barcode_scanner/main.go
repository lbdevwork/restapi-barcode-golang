package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/api"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

var database *sql.DB

func main() {

	var err2 error
	// Create a connection to the database
	database, err2 = connectToDatabase()
	if err2 != nil {
		log.Fatalf("Failed to open database connection: %v\n", err2)
	}
	defer database.Close()

	// Checks if tables exist and creates them if they don't
	checkTables()

	// Define the routes for accessing the API
	router := setupRouter()

	// Define the port to listen on
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	// Console output of API status
	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}

// Verify that the .env file exists and can be acessed
func checkEnvFile() {

	// Get the absolute path of the .env file
	_, currentFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(currentFile)
	rootPath := filepath.Join(basePath, "..", "..")
	envPath := filepath.Join(rootPath, ".env")

	// Load environment variables from .env file
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}
}

// Checks if tables exist and creates them if they don't
func checkTables() {
	err := db.CreateSchema(database)
	if err != nil {
		log.Fatalf("Error creating database schema: %v\n", err)
	}
}

// Define the routes for accessing the API
func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/product/{barcode}", productHandler)
		r.Get("/product/text/{lang}/{barcode}", productHandlerText)
	})
	return r
}

// Fetches a product from the database or the API
func fetchProduct(w http.ResponseWriter, r *http.Request, handleResponse func(http.ResponseWriter, db.Product)) {
	barcode := chi.URLParam(r, "barcode")

	barcode = utils.ConvertTo13DigitNumber(barcode)

	if barcode == "" || barcode == "error" {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	product, err := db.FetchProduct(ctx, database, barcode)

	if err != nil {
		log.Printf("Product not found in local database: %s\n", barcode)

		product, err = api.FetchProduct(ctx, barcode)
		if err != nil {
			log.Printf("Failed to fetch product from external API: %v\n", err)
			return
		}

		if product.ID == "" {
			log.Printf("Product ID is empty after fetching from external API\n")
			return
		}

		err = db.StoreProduct(ctx, database, product)
		if err != nil {
			log.Printf("Failed to store product in local database: %v\n", err)
			return
		}
	} else {
		fmt.Printf("Product found in local database: %s\n", barcode)
	}

	handleResponse(w, product)
}

// Handles Json Endpoint Response
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

// Handles Text Endpoint Response
func productHandlerText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fetchProduct(w, r, func(w http.ResponseWriter, product db.Product) {
		if product.ID == "" {
			handleError(w, http.StatusNotFound, "Barcode not found", nil)
			return
		}
		language := chi.URLParam(r, "lang")
		if language == "en" {

			fmt.Fprintf(w, "Barcode: %s\nProduct Name: %s\nNutriscore Grade: %s (From A to E) (Representative Category of the Product's Nutricional Value)\nEcoscore Grade: %s (From A to E) (Representative Category of the Product's Carbon Footprint)\n\nPer 100g:\n - Energy: %f kJ (%f kcal)\n - Fat: %f g\n - Saturated Fat: %f g\n - Carbohydrates: %f g\n - Sugars: %f g\n - Protein: %f g\n - Fiber: %f g\n - Salt: %f g\n - Sodium: %f g",
				product.ID, product.ProductName, product.NutriscoreGrade, product.EcoscoreGrade,
				product.Nutriments.EnergyKJ, product.Nutriments.EnergyKcal,
				product.Nutriments.Fat, product.Nutriments.SaturatedFat,
				product.Nutriments.Carbohydrates, product.Nutriments.Sugars,
				product.Nutriments.Protein, product.Nutriments.Fiber,
				product.Nutriments.Salt, product.Nutriments.Sodium)
		} else if language == "pt" {

			fmt.Fprintf(w, "Código de barras: %s\nNome do Produto: %s\nClassificação Nutriscore: %s (De A a E) (Categoria Representativa do Valor Nutricional do Produto)\nClassificação Ecoscore: %s (De A a E) (Categoria Representativa da Pegada de Carbono do Produto)\n\nPor 100g:\n - Energia: %f kJ (%f kcal)\n - Gordura: %f g\n - Gordura Saturada: %f g\n - Carboidratos: %f g\n - Açúcares: %f g\n - Proteínas: %f g\n - Fibras: %f g\n - Sal: %f g\n - Sódio: %f g",
				product.ID, product.ProductName, product.NutriscoreGrade, product.EcoscoreGrade,
				product.Nutriments.EnergyKJ, product.Nutriments.EnergyKcal,
				product.Nutriments.Fat, product.Nutriments.SaturatedFat,
				product.Nutriments.Carbohydrates, product.Nutriments.Sugars,
				product.Nutriments.Protein, product.Nutriments.Fiber,
				product.Nutriments.Salt, product.Nutriments.Sodium)
		}
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

func connectToDatabase() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error: %s environment variable not set.", k)
		}
		return v
	}

	var (
		dbUser           = mustGetenv("DB_USER")
		dbPwd            = mustGetenv("DB_PASS")
		dbName           = mustGetenv("DB_NAME")
		instanceConnName = mustGetenv("INSTANCE_CONNECTION_NAME")
	)

	socketDir := "/cloudsql"

	dbURI := fmt.Sprintf("%s:%s@unix(%s/%s)/%s?parseTime=true",
		dbUser, dbPwd, socketDir, instanceConnName, dbName)

	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	return db, nil
}
