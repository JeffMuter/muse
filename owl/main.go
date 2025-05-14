package owl

import (
	"database/sql"
	"log"
)

func main() {

	db, err := initDatabase()
	if err != nil {
		log.Fatal("failed to initialize database... %v\n", err)
	}

}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "../db/muse.db")
	if err != nil {
		return nil, err
	}

	// Run migrations or schema setup here if needed
	// You might want to execute the schema.sql file here

	return db, nil
}
