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
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/search?code=%s&fields=code,product_name,ecoscore_grade,nutriscore_grade", barcode)

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

return product, nil
}


