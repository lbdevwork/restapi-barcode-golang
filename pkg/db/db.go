package db

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID                 string `json:"id"`
	ProductName        string `json:"product_name"`
	ProductNameEn      string `json:"product_name_en"`
	ProductQuantity    string `json:"product_quantity"`
	Quantity           string `json:"quantity"`
	ServingQuantity    string `json:"serving_quantity"`
	ServingSize        string `json:"serving_size"`
	Ingredients        string `json:"ingredients_text_en"`
	NutritionGradeFr   string `json:"nutrition_grade_fr"`
	NutritionDataPer   string `json:"nutrition_data_per"`
	Categories         string `json:"categories"`
	CategoriesTags     string `json:"categories_tags"`
	Brands             string `json:"brands"`
	BrandsTags         string `json:"brands_tags"`
	Traces             string `json:"traces"`
	TracesTags         string `json:"traces_tags"`
	Countries          string `json:"countries"`
	CountriesTags      string `json:"countries_tags"`
	PurchasePlacesTags string `json:"purchase_places_tags"`
	StoresTags         string `json:"stores_tags"`
}

func FetchProduct(ctx context.Context, db *sql.DB, barcode string) (Product, error) {
	var product Product
	err := db.QueryRowContext(ctx, "SELECT id, product_name, product_name_en, product_quantity, quantity, serving_quantity, serving_size, ingredients_text_en, nutrition_grade_fr, nutrition_data_per, categories, categories_tags, brands, brands_tags, traces, traces_tags, countries, countries_tags, purchase_places_tags, stores_tags FROM products WHERE id = ?", barcode).Scan(&product.ID, &product.ProductName, &product.ProductNameEn, &product.ProductQuantity, &product.Quantity, &product.ServingQuantity, &product.ServingSize, &product.Ingredients, &product.NutritionGradeFr, &product.NutritionDataPer, &product.Categories, &product.CategoriesTags, &product.Brands, &product.BrandsTags, &product.Traces, &product.TracesTags, &product.Countries, &product.CountriesTags, &product.PurchasePlacesTags, &product.StoresTags)

	return product, err
}

func FetchProductID(ctx context.Context, db *sql.DB, productID string) (string, error) {
	var id string
	err := db.QueryRowContext(ctx, "SELECT id FROM products WHERE id = ?", productID).Scan(&id)
	return id, err
}

func StoreProduct(ctx context.Context, db *sql.DB, product Product) error {
	_, err := db.ExecContext(ctx, `INSERT INTO products (id, product_name, product_name_en, product_quantity, quantity, serving_quantity, serving_size, ingredients_text_en, nutrition_grade_fr, nutrition_data_per, categories, categories_tags, brands, brands_tags, traces, traces_tags, countries, countries_tags, purchase_places_tags, stores_tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		product.ID,
		product.ProductName,
		product.ProductNameEn,
		product.ProductQuantity,
		product.Quantity,
		product.ServingQuantity,
		product.ServingSize,
		product.Ingredients,
		product.NutritionGradeFr,
		product.NutritionDataPer,
		product.Categories,
		product.CategoriesTags,
		product.Brands,
		product.BrandsTags,
		product.Traces,
		product.TracesTags,
		product.Countries,
		product.CountriesTags,
		product.PurchasePlacesTags,
		product.StoresTags,
	)
	return err
}

func safeString(value interface{}) string {
	if value != nil {
		return value.(string)
	}
	return ""
}

/* product := db.Product{
    ID:                 safeString(productData["id"]),
    ProductName:        safeString(productData["product_name"]),
    ProductNameEn:      safeString(productData["product_name_en"]),
    ProductQuantity:    safeString(productData["product_quantity"]),
    Quantity:           safeString(productData["quantity"]),
    ServingQuantity:    safeString(productData["serving_quantity"]),
    ServingSize:        safeString(productData["serving_size"]),
    Ingredients:        safeString(productData["ingredients_text_en"]),
    NutritionGradeFr:   safeString(productData["nutrition_grade_fr"]),
    NutritionDataPer:   safeString(productData["nutrition_data_per"]),
    Categories:         safeString(productData["categories"]),
    CategoriesTags:     safeString(productData["categories_tags"]),
    Brands:             safeString(productData["brands"]),
    BrandsTags:         safeString(productData["brands_tags"]),
    Traces:             safeString(productData["traces"]),
    TracesTags:         safeString(productData["traces_tags"]),
    Countries:          safeString(productData["countries"]),
    CountriesTags:      safeString(productData["countries_tags"]),
    PurchasePlacesTags: safeString(productData["purchase_places_tags"]),
    StoresTags:         safeString(productData["stores_tags"]),
}*/
