package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lbdevwork/restapi-barcode-golang/pkg/db"

	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

func FetchProduct(ctx context.Context, barcode string) (db.Product, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/search?code=%s&fields=code,product_name,ecoscore_grade,nutriscore_grade,nutriments", barcode)

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

	products, ok := data["products"].([]interface{})
	if !ok || len(products) == 0 {
		return db.Product{}, fmt.Errorf("unable to find the products key in the response data or products array is empty")
	}

	productData := products[0].(map[string]interface{})
	product := db.Product{
		ID:              productData["code"].(string),
		ProductName:     utils.SafeString(productData["product_name"]),
		NutriscoreGrade: utils.SafeString(productData["nutriscore_grade"]),
		EcoscoreGrade:   utils.SafeString(productData["ecoscore_grade"]),
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

	nutrimentsData := productData["nutriments"].(map[string]interface{})

	//Calculate energy-kj_100g and energy-kcal_100g from energy_100g if needed
	if utils.SafeFloat64(nutrimentsData["energy-kj_100g"]) == 0 && utils.SafeFloat64(nutrimentsData["energy-kcal_100g"]) != 0 {
		nutrimentsData["energy-kj_100g"] = utils.SafeFloat64(nutrimentsData["energy-kcal_100g"]) * 4.184
	} else if utils.SafeFloat64(nutrimentsData["energy-kj_100g"]) != 0 && utils.SafeFloat64(nutrimentsData["energy-kcal_100g"]) == 0 {
		nutrimentsData["energy-kcal_100g"] = utils.SafeFloat64(nutrimentsData["energy-kj_100g"]) / 4.184
	}

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
