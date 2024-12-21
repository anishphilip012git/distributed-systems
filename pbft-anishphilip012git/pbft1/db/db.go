package db

import (
	"fmt"
	"log"
	"os"

	"go.etcd.io/bbolt"
)

// Global variable for database
var DB *bbolt.DB

// Initialize BoltDB
func InitBoltDB(dbPath string) (*bbolt.DB, error) {
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open BoltDB: %w", err)
	}
	return db, nil
}

// Clear the BoltDB file, check if the file exists first
func ClearDB(path string) error {
	// Check if the file exists before trying to remove it
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		log.Printf("Database at %s cleared successfully.", path)
	} else if os.IsNotExist(err) {
		log.Printf("Database file %s does not exist, skipping removal.", path)
	} else {
		return fmt.Errorf("error checking file existence: %w", err)
	}
	return nil
}

// Initialize database connections map
var DbConnections map[string]*bbolt.DB
var (
	TBL_EXEC = "EXEC"
)

// Start a BoltDB instance for a specific server
func StartDB(server string) *bbolt.DB {
	// Ensure the directory exists before creating the file
	dir := "./bolt"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Set database path
	path := dir + "/" + server + ".db"
	err := ClearDB(path)
	if err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}
	log.Println("Starting BoltDB for server:", server, "Path:", path)

	// Open or create the database
	db, err := InitBoltDB(path)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	return db
}

// Start all BoltDB instances for each server
func StartallDB(svrs []string) {
	DbConnections = map[string]*bbolt.DB{}
	for _, svr := range svrs {
		DbConnections[svr] = StartDB(svr)
	}
}
func ClearAllDB(svrs []string) {
	for _, svr := range svrs {
		log.Println("DB clearing start", svr, DbConnections[svr])
		DeleteAllRecords(DbConnections[svr], TBL_EXEC)
		log.Printf("DB cleared successfully for %s", svr)
	}
}

// func main() {
// 	// Example usage
// 	servers := []string{"server1", "server2", "server3"}
// 	StartallDB(servers)
// }
