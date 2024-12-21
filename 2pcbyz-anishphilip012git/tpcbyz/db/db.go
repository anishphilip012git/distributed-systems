package db

import (
	"fmt"
	"log"
	"os"
	"time"
	"tpcbyz/metric"

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

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(TBL_WAL))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(TBL_EXEC))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
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
var DBBalMap map[string]*ConcurrentBoltMap
var (
	TBL_EXEC     = "EXEC"
	TBL_WAL      = "WAL"
	TBL_TXN_TIME = "TXN"
	TBL_BALANCE  = "BAL"
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
	// walPath := path + ".wal"
	// b_map, _ := NewConcurrentBoltMap(db, walPath)
	// DBBalMap[server] = b_map
	return db
}

// Start all BoltDB instances for each server
func StartallDB(svrs []string) {
	DbConnections = map[string]*bbolt.DB{}
	for _, svr := range svrs {
		DbConnections[svr] = StartDB(svr)
	}
}

func GetTxnTimes(db *bbolt.DB) {
	recs, err := ScanTable(db, TBL_TXN_TIME)
	if err != nil {
		log.Printf("err %v", err)

	}
	output := ""
	for index, txn := range recs {
		output += fmt.Sprintf("[%d: %s] , ", index, txn)
	}
	log.Println("Time to Process Each Txn:", output)
	avgLatency := metric.TnxSum.TotalElapsedTime / time.Duration(metric.TnxSum.TransactionCount)
	log.Printf("Avg Time: %v ", avgLatency)
}

// func main() {
// 	// Example usage
// 	servers := []string{"server1", "server2", "server3"}
// 	StartallDB(servers)
// }
