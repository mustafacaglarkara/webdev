package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Migrate: .up.sql dosyalarını çalıştırır
func Migrate(db *sql.DB, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations klasörü okunamadı: %w", err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".up.sql" {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, f.Name()))
			if err != nil {
				return fmt.Errorf("dosya okunamadı: %w", err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("göç hatası (%s): %w", f.Name(), err)
			}
		}
	}
	return nil
}

// Rollback: .down.sql dosyalarını çalıştırır
func RollbackMigrations(db *sql.DB, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations klasörü okunamadı: %w", err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".down.sql" {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, f.Name()))
			if err != nil {
				return fmt.Errorf("dosya okunamadı: %w", err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("rollback hatası (%s): %w", f.Name(), err)
			}
		}
	}
	return nil
}
