package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	// Enable WAL mode in the connection string
	dsn := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&cache=shared"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Double check WAL mode
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, err
	}

	// Run migrations automatically
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrationFile := filepath.Join("migrations", "001_init.sql")
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Printf("Failed to read migration file %s: %v", migrationFile, err)
		return err
	}

	_, err = db.Exec(string(content))
	if err != nil {
		log.Printf("Failed to execute migration: %v", err)
		return err
	}

	log.Println("Migrations executed successfully")
	return nil
}
