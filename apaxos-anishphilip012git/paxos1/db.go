package main

import (
	"fmt"
	"log"
	"os"
	pb "paxos1/proto"
	"time"

	"github.com/dgraph-io/badger/v4"
)

func InitBadgerDB(dbPath string) (*badger.DB, error) {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.INFO)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}
	return db, nil
}

// Clear the BadgerDB directory
func ClearDB(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	log.Printf("Database at %s cleared successfully.", path)
	return nil
}

func StartDB() *badger.DB {
	path := "./badger"
	err := ClearDB(path)
	if err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}

	db, err := InitBadgerDB(path)
	if err != nil {
		log.Fatal(err)
	}
	return db

}

func StoreRecord(db *badger.DB, tableName string, index int64, transaction *pb.Transaction) error {
	// Generate a table-specific key: <table_name>:<key>
	value := fmt.Sprintf("%s|%s|%d", transaction.Sender, transaction.Receiver, transaction.Amount)
	fullKey := fmt.Sprintf("%s:%d", tableName, index)
	// log.Printf("Updating record with key:%s and value:%s \n", fullKey, value)
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fullKey), []byte(value))
	})
}

func StoreTime(db *badger.DB, index int64, transaction *pb.Transaction, timeTaken time.Duration) error {
	// Generate a table-specific key: <table_name>:<key>
	value := fmt.Sprintf("%s|%s|%d- took %v ", transaction.Sender, transaction.Receiver, transaction.Amount, timeTaken)
	fullKey := fmt.Sprintf("%s:%d", "txn", index)
	// log.Printf("Updating record with key:%s and value:%s \n", fullKey, value)
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fullKey), []byte(value))
	})
}

/*

record, err := GetRecord(db, "users", "1")
if err != nil {
	log.Fatalf("Failed to get user: %v", err)
} else {
	log.Println("Retrieved user:", record)
}

*/

func GetRecord(db *badger.DB, tableName string, index int64) (string, error) {
	fullKey := fmt.Sprintf("%s:%d", tableName, index)

	var value string
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fullKey))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
	})

	if err != nil {
		return "", err
	}
	return value, nil
}

// Get all records from the table with an index greater than the given index
func GetRecordsAfter(db *badger.DB, tableName string, minIndex int64) (map[int64]string, error) {
	records := make(map[int64]string)

	err := db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("%s:", tableName))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Extract the index from the key
			var index int64
			if _, err := fmt.Sscanf(key, fmt.Sprintf("%s:%%d", tableName), &index); err != nil {
				return err
			}

			// Filter entries with index > minIndex
			if index > minIndex {
				err := item.Value(func(val []byte) error {
					records[index] = string(val)
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return records, nil
}

/*
users, err := ScanTable(db, "users")

	if err != nil {
		log.Fatalf("Failed to scan users table: %v", err)
	} else {

		log.Println("Users:", users)
	}
*/
func ScanTable(db *badger.DB, tableName string) ([]string, error) {
	var records []string

	prefix := []byte(fmt.Sprintf("%s:", tableName))
	err := db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			err := item.Value(func(val []byte) error {
				records = append(records, string(val))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return records, nil
}
func GetTxnTimes() {
	recs, err := ScanTable(DB, "txn")
	if err != nil {
		log.Printf("err %v", err)

	}
	output := ""
	for index, txn := range recs {
		output += fmt.Sprintf("[%d: %s] , ", index, txn)
	}
	log.Println("Time to Process Each Txn:", output)
	avgLatency := TnxSum.TotalElapsedTime / time.Duration(TnxSum.TransactionCount)
	log.Printf("Avg Time: %v ", avgLatency)
}

/*
err := DeleteRecord(db, "users", "1")
if err != nil {
	log.Fatalf("Failed to delete user: %v", err)
}
*/

func DeleteRecord(db *badger.DB, tableName string, key string) error {
	fullKey := fmt.Sprintf("%s:%s", tableName, key)

	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(fullKey))
	})
}
