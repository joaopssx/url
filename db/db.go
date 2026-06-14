package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	dsn := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&cache=shared"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{"001_init.sql", "002_add_webhook.sql"}
	for _, m := range migrations {
		migrationFile := filepath.Join("migrations", m)
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			log.Printf("Failed to read migration file %s: %v", migrationFile, err)
			return err
		}

		_, err = db.Exec(string(content))
		if err != nil {
			if m == "002_add_webhook.sql" && strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			log.Printf("Failed to execute migration: %v", err)
			return err
		}
	}
	return nil
}
