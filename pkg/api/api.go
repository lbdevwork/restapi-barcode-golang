package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/certifi/gocertifi"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"
	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

func FetchProduct(ctx context.Context, barcode string) (db.Product, error) {
	httpClient, err := createCustomHTTPClient()
	if err != nil {
		return db.Product{}, fmt.Errorf("error creating custom HTTP client: %s", err)
	}

	productDataBytes, err := fetchProductDataJson(ctx, httpClient, barcode)
	if err != nil {
		return db.Product{}, fmt.Errorf("error fetching product data: %s", err)
	}
	// Decompress the JSON to a map
	var productDataMap map[string]interface{}
	err = json.Unmarshal(productDataBytes, &productDataMap)
	if err != nil {
		return db.Product{}, fmt.Errorf("error unmarshalling product data JSON: %s", err)
	}

	// Extract values from the decompressed JSON data
	product := db.Product{
		ID:              utils.SafeString(productDataMap["code"]),
		ProductName:     utils.SafeString(productDataMap["product_name"]),
		NutriscoreGrade: utils.SafeString(productDataMap["nutriscore_grade"]),
		EcoscoreGrade:   utils.SafeString(productDataMap["ecoscore_grade"]),
	}

	if product.ProductName == "" {
		product.ProductName = "unknown"
	}
	if product.NutriscoreGrade == "" {
		product.NutriscoreGrade = "unknown"
	}
	if product.EcoscoreGrade == "" {
		product.EcoscoreGrade = "unknown"
	}

	nutrimentsData, ok := productDataMap["nutriments"].(map[string]interface{})
	if !ok {
		return db.Product{}, fmt.Errorf("error extracting nutriments data")
	}
	fmt.Println("DEBUG: nutrimentsData:", nutrimentsData)
	fmt.Println("DEBUG: fat_100g:", nutrimentsData["fat_100g"])
	fmt.Println("DEBUG: SafeFloat64 fat_100g:", utils.SafeFloat64(nutrimentsData["fat_100g"]))
	product.Nutriments = db.Nutriments{
		EnergyKJ:      utils.SafeFloat64(nutrimentsData["energy-kj_100g"]),
		EnergyKcal:    utils.SafeFloat64(nutrimentsData["energy-kcal_100g"]),
		Fat:           utils.SafeFloat64(nutrimentsData["fat_100g"]),
		SaturatedFat:  utils.SafeFloat64(nutrimentsData["saturated-fat_100g"]),
		Carbohydrates: utils.SafeFloat64(nutrimentsData["carbohydrates_100g"]),
		Sugars:        utils.SafeFloat64(nutrimentsData["sugars_100g"]),
		Protein:       utils.SafeFloat64(nutrimentsData["proteins_100g"]),
		Fiber:         utils.SafeFloat64(nutrimentsData["fiber_100g"]),
		Salt:          utils.SafeFloat64(nutrimentsData["salt_100g"]),
		Sodium:        utils.SafeFloat64(nutrimentsData["sodium_100g"]),
	}

	return product, nil
}

func fetchProductDataJson(ctx context.Context, httpClient *http.Client, barcode string) ([]byte, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/search?code=%s&fields=code,product_name,ecoscore_grade,nutriscore_grade,nutriments", barcode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the request: %s", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching the URL: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching the product data: non-200 status code (%d)", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response body: %s", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling the JSON data: %s", err)
	}

	products, ok := data["products"].([]interface{})
	if !ok || len(products) == 0 {
		return nil, fmt.Errorf("unable to find the products key in the response data or products array is empty")
	}

	status, ok := data["status"].(float64)
	if ok && status != 1 {
		return nil, fmt.Errorf("unable to find the status key in the response data or the product was not found")
	}

	productData := products[0].(map[string]interface{})
	productDataBytes, err := json.Marshal(productData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the product data: %s", err)
	}

	return productDataBytes, nil
}

func createCustomHTTPClient() (*http.Client, error) {
	caCertPool, err := gocertifi.CACerts()
	if err != nil {
		return nil, fmt.Errorf("failed to load gocertifi certificate pool: %s", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}, nil
}
