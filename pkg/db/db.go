package db

import (
	"context"
	"database/sql"
	"fmt"
	
	_ "github.com/go-sql-driver/mysql"

	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

type Product struct {
	ID              string
	ProductName     string
	NutriscoreGrade string
	EcoscoreGrade   string
}

func FetchProduct(ctx context.Context, db *sql.DB, barcode string) (Product, error) {
	var product Product

	err := db.QueryRowContext(ctx, "SELECT id, product_name, nutriscore_grade, ecoscore_grade FROM products WHERE id = ?", barcode).Scan(&product.ID, &product.ProductName, &product.NutriscoreGrade, &product.EcoscoreGrade)

	// Add this block of code to log the error and see what's going wrong
	if err != nil {
		fmt.Printf("Error fetching product from the database: %v\n", err)
	}

	return product, err
}

func StoreProduct(ctx context.Context, db *sql.DB, product Product) error {
	product.ID = utils.convertTo13DigitNumber(product.ID)
	_, err := db.ExecContext(ctx, `INSERT IGNORE INTO products (id, product_name, nutriscore_grade, ecoscore_grade) VALUES (?, ?, ?, ?)`,
		product.ID,
		product.ProductName,
		product.NutriscoreGrade,
		product.EcoscoreGrade,
	)
	return err
}


func CreateSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			product_name TEXT,
			nutriscore_grade TEXT,
			ecoscore_grade TEXT
		);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
