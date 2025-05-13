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
	owlService := service.NewOwlService(db)
	port := "50052" // Different from pigeon's port

	// Start gRPC server
	if err := server.StartGRPCServer(port, owlService); err != nil {
		log.Fatal("Failed to start gRPC server:", err)
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
