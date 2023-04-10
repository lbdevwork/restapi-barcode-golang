package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/lbdevwork/restapi-barcode-golang/pkg/utils"
)

type Product struct {
	ID              string     `json:"id"`
	ProductName     string     `json:"product_name"`
	NutriscoreGrade string     `json:"nutriscore_grade"`
	EcoscoreGrade   string     `json:"ecoscore_grade"`
	Nutriments      Nutriments `json:"nutriments"`
}

type Nutriments struct {
	EnergyKJ      float64
	EnergyKcal    float64
	Fat           float64
	SaturatedFat  float64
	Carbohydrates float64
	Sugars        float64
	Protein       float64
	Fiber         float64
	Salt          float64
	Sodium        float64
}

type ProductNotFoundError struct {
	Barcode string
}

func FetchProduct(ctx context.Context, db *sql.DB, barcode string) (Product, error) {
	var product Product
	var nutriments Nutriments

	err := db.QueryRowContext(ctx, `
        SELECT p.id, p.product_name, p.nutriscore_grade, p.ecoscore_grade, 
               n.energy_kj, n.energy_kcal, n.fat, n.saturated_fat, n.carbohydrates, n.sugars, 
               n.protein, n.fiber, n.salt, n.sodium
        FROM products p
        LEFT JOIN nutriments n ON p.id = n.product_id
        WHERE p.id = ?`, barcode).Scan(
		&product.ID,
		&product.ProductName,
		&product.NutriscoreGrade,
		&product.EcoscoreGrade,
		&nutriments.EnergyKJ,
		&nutriments.EnergyKcal,
		&nutriments.Fat,
		&nutriments.SaturatedFat,
		&nutriments.Carbohydrates,
		&nutriments.Sugars,
		&nutriments.Protein,
		&nutriments.Fiber,
		&nutriments.Salt,
		&nutriments.Sodium,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Product{}, &ProductNotFoundError{Barcode: barcode}
		}
		return Product{}, err
	}

	product.Nutriments = nutriments
	return product, nil
}

func StoreProduct(ctx context.Context, db *sql.DB, product Product) error {
	product.ID = utils.ConvertTo13DigitNumber(product.ID)
	if product.ID == "error" {
		return fmt.Errorf("invalid barcode: %s", product.ID)
	}
	productInsert := `
		INSERT IGNORE INTO products (id, product_name, nutriscore_grade, ecoscore_grade)
		VALUES (?, ?, ?, ?);
	`
	nutrimentsInsert := `
		INSERT IGNORE INTO nutriments (product_id, energy_kj, energy_kcal, fat, saturated_fat, carbohydrates, sugars, protein, fiber, salt, sodium)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, productInsert, product.ID, product.ProductName, product.NutriscoreGrade, product.EcoscoreGrade)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, nutrimentsInsert, product.ID, product.Nutriments.EnergyKJ, product.Nutriments.EnergyKcal, product.Nutriments.Fat, product.Nutriments.SaturatedFat, product.Nutriments.Carbohydrates, product.Nutriments.Sugars, product.Nutriments.Protein, product.Nutriments.Fiber, product.Nutriments.Salt, product.Nutriments.Sodium)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func CreateSchema(db *sql.DB) error {

	productsSchema := `
		CREATE TABLE IF NOT EXISTS products (
			id VARCHAR(13) PRIMARY KEY,
			product_name VARCHAR(255),
			nutriscore_grade VARCHAR(255),
			ecoscore_grade VARCHAR(255)
		);
	`

	nutrimentsSchema := `
		CREATE TABLE IF NOT EXISTS nutriments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			product_id VARCHAR(13) NOT NULL,
			energy_kj DECIMAL(10, 2),
			energy_kcal DECIMAL(10, 2),
			fat DECIMAL(10, 2),
			saturated_fat DECIMAL(10, 2),
			carbohydrates DECIMAL(10, 2),
			sugars DECIMAL(10, 2),
			protein DECIMAL(10, 2),
			fiber DECIMAL(10, 2),
			salt DECIMAL(10, 2),
			sodium DECIMAL(10, 2),
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
		);
	`

	_, err := db.Exec(productsSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(nutrimentsSchema)
	if err != nil {
		return err
	}

	return nil
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("Product with barcode %s not found", e.Barcode)
}
