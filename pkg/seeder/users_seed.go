package seeder

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// MigrationRunner: SQL dosyalarını bulup çalıştırır
func RunMigrations(db *sql.DB, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations klasörü okunamadı: %w", err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, f.Name()))
			if err != nil {
				return fmt.Errorf("dosya okunamadı: %w", err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("migration hatası (%s): %w", f.Name(), err)
			}
		}
	}
	return nil
}

// Rollback için benzer fonksiyon eklenebilir
