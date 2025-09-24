package seeder

import (
	"database/sql"
	"fmt"
)

// SeedRunner: seed fonksiyonlarını çalıştırır
func RunSeeds(db *sql.DB, seeds ...func(*sql.DB) error) error {
	for _, seed := range seeds {
		if err := seed(db); err != nil {
			return fmt.Errorf("seed hatası: %w", err)
		}
	}
	return nil
}
