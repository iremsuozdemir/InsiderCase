package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect establishes a connection to PostgreSQL
func Connect() error {
	// Connection string for PostgreSQL 15 on port 5433
	connectionString := "host=localhost port=5433 user=iremsuozdemir dbname=football_league sslmode=disable"
	
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	
	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	
	fmt.Println("✅ Connected to PostgreSQL database successfully!")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
} 