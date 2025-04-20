package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DBInterface defines a contract for working with the database
type DBInterface interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Begin() (*sql.Tx, error)
}

// BD implements DBInterface
type DB struct {
	*sql.DB
}

// Query implements the method Query
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// QueryRow implements the method QueryRow
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

// Exec implements the method Exec
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

func (db *DB) Begin() (*sql.Tx, error) {
	return db.DB.Begin()
}

// InitDB connects to the database
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return nil, err
	}

	// clearTables(db)
	createTables(db)

	return db, nil
}

func clearTables(db *sql.DB) {
	tables := []string{"users", "pvz", "receptions", "products"}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			log.Printf("Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Table %s dropped successfully", table)
		}
	}
}

// createTables creates a database
func createTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id TEXT NOT NULL PRIMARY KEY,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            role TEXT NOT NULL CHECK (role IN ('employee', 'moderator'))
        )`,
		`CREATE TABLE IF NOT EXISTS pvz (
            id TEXT NOT NULL PRIMARY KEY,
            city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
            registration_date TEXT NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS receptions (
    		id TEXT NOT NULL PRIMARY KEY,
    		date_time TEXT NOT NULL,
    		pvz_id TEXT NOT NULL,
			product_ids TEXT DEFAULT '[]',
    		status TEXT NOT NULL CHECK (status IN ('in_progress', 'close')),
    		FOREIGN KEY (pvz_id) REFERENCES pvz(id)
		)`,
		`CREATE TABLE IF NOT EXISTS products (
    		id TEXT NOT NULL PRIMARY KEY,
    		date_time TEXT NOT NULL,
    		type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_id ON products(id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_receptions_pvz_status ON receptions(pvz_id, status) WHERE status = 'in_progress'`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Failed to execute query: %v", err)
		}
	}
}
