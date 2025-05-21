package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	// use db instead of _ after testing
	_, err := initDatabase()
	if err != nil {
		log.Fatalf("failed to initialize database... %v\n", err)
	}
	fmt.Println("db initialized successfully...")
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./db/muse.db")
	if err != nil {
		return nil, err
	}

	// Run migrations or schema setup here if needed
	// You might want to execute the schema.sql file here

	return db, nil
}
