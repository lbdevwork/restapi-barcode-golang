package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
)

func FetchProduct(ctx context.Context, barcode string) (db.Product, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return db.Product{}, fmt.Errorf("error creating the request: %s", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return db.Product{}, fmt.Errorf("error fetching the URL: %s", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return db.Product{}, fmt.Errorf("error reading the response body: %s", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return db.Product{}, fmt.Errorf("error unmarshalling the JSON data: %s", err)
	}

	productData := data["product"].(map[string]interface{})
	product := db.Product{
		ID:                 productData["id"].(string),
		ProductName:        productData["product_name"].(string),
		ProductNameEn:      productData["product_name_en"].(string),
		ProductQuantity:    productData["product_quantity"].(string),
		Quantity:           productData["quantity"].(string),
		ServingQuantity:    productData["serving_quantity"].(string),
		ServingSize:        productData["serving_size"].(string),
		Ingredients:        productData["ingredients_text_en"].(string),
		NutritionGradeFr:   productData["nutrition_grade_fr"].(string),
		NutritionDataPer:   productData["nutrition_data_per"].(string),
		Categories:         productData["categories"].(string),
		CategoriesTags:     sliceToString(productData["categories_tags"]),
		Brands:             productData["brands"].(string),
		BrandsTags:         sliceToString(productData["brands_tags"]),
		Traces:             productData["traces"].(string),
		TracesTags:         sliceToString(productData["traces_tags"]),
		Countries:          productData["countries"].(string),
		CountriesTags:      sliceToString(productData["countries_tags"]),
		PurchasePlacesTags: sliceToString(productData["purchase_places_tags"]),
		StoresTags:         sliceToString(productData["stores_tags"]),
	}

	return product, nil
}

func sliceToString(v interface{}) string {
	if slice, ok := v.([]interface{}); ok {
		strs := make([]string, len(slice))
		for i, s := range slice {
			strs[i] = s.(string)
		}
		return strings.Join(strs, ",")
	}
	return ""
}

/*
func FetchProduct2(ctx context.Context, barcode string) (db.ProductInfo, error) {
	apiURL := fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return db.ProductInfo{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return db.ProductInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return db.ProductInfo{}, fmt.Printf("Failed to fetch product information")
	}

	var apiResponse struct {
		Status        int    `json:"status"`
		StatusVerbose string `json:"status_verbose"`
		Product       struct {
			Name            string  `json:"product_name"`
			Nutriscore      string  `json:"nutriscore_grade"`
			CarbonFootprint float64 `json:"carbon_footprint"`
		} `json:"product"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return db.ProductInfo{}, err
	}

	return db.ProductInfo{
		Barcode:         barcode,
		Name:            apiResponse.Product.Name,
		Nutriscore:      apiResponse.Product.Nutriscore,
		CarbonFootprint: apiResponse.Product.CarbonFootprint,
	}, nil
}*/
