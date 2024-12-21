package db

import (
	"fmt"
	"time"
	pb "tpcbyz/proto"

	"go.etcd.io/bbolt"
)

// StoreRecord stores a transaction record in a specific bucket in BoltDB
func StoreRecord(db *bbolt.DB, tableName string, index int64, transaction *pb.Transaction) error {
	// Open the bucket or create it if it doesn't exist
	err := db.Update(func(txn *bbolt.Tx) error {
		bucket, err := txn.CreateBucketIfNotExists([]byte(tableName))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// Generate the record's value
		value := fmt.Sprintf("%d::%d=>%d|%d|%d", index, transaction.SeqNo, transaction.Sender, transaction.Receiver, transaction.Amount)
		fullKey := fmt.Sprintf("%d", index)

		// Store the record in the bucket
		return bucket.Put([]byte(fullKey), []byte(value))
	})

	return err
}

// StoreTime stores the time taken for a transaction in the "txn" bucket
func StoreTime(db *bbolt.DB, index int64, transaction *pb.Transaction, timeTaken time.Duration) error {
	// Open the bucket or create it if it doesn't exist
	err := db.Update(func(txn *bbolt.Tx) error {
		bucket, err := txn.CreateBucketIfNotExists([]byte(TBL_TXN_TIME))
		if err != nil {
			return fmt.Errorf("failed to create 'txn' bucket: %w", err)
		}

		// Generate the record's value
		value := fmt.Sprintf("%d::%d|%d|%d - took %v ", index, transaction.Sender, transaction.Receiver, transaction.Amount, timeTaken)
		fullKey := fmt.Sprintf("%d", index)

		// Store the record in the "txn" bucket
		return bucket.Put([]byte(fullKey), []byte(value))
	})

	return err
}

// GetRecord retrieves a record by its index from a specific bucket in BoltDB
func GetRecord(db *bbolt.DB, tableName string, index int64) (string, error) {
	var value string
	err := db.View(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(tableName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", tableName)
		}

		fullKey := fmt.Sprintf("%d", index)
		val := bucket.Get([]byte(fullKey))
		if val == nil {
			return fmt.Errorf("record not found for key %s", fullKey)
		}

		value = string(val)
		return nil
	})

	if err != nil {
		return "", err
	}
	return value, nil
}

// GetRecordsAfter retrieves all records with an index greater than the given minIndex from a specific bucket in BoltDB
func GetRecordsAfter(db *bbolt.DB, tableName string, minIndex int64) (map[int64]string, error) {
	records := make(map[int64]string)

	err := db.View(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(tableName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", tableName)
		}

		// Iterate over all the keys in the bucket
		cursor := bucket.Cursor()
		for k, v := cursor.Seek([]byte(fmt.Sprintf("%d", minIndex+1))); k != nil; k, v = cursor.Next() {
			// Convert key (index) to int64
			var index int64
			_, err := fmt.Sscanf(string(k), "%d", &index)
			if err != nil {
				return fmt.Errorf("failed to parse key %s as int64: %w", k, err)
			}

			// Store the record if the index is greater than minIndex
			records[index] = string(v)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return records, nil
}

// ScanTable retrieves all records from a specific bucket in BoltDB
func ScanTable(db *bbolt.DB, tableName string) ([]string, error) {
	var records []string

	err := db.View(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(tableName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", tableName)
		}

		// Iterate over all the keys in the bucket
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			records = append(records, string(v))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return records, nil
}

// DeleteRecord deletes a record by its key from a specific bucket in BoltDB
func DeleteRecord(db *bbolt.DB, tableName string, index int64) error {
	err := db.Update(func(txn *bbolt.Tx) error {
		bucket := txn.Bucket([]byte(tableName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", tableName)
		}
		fullKey := fmt.Sprintf("%d", index)
		return bucket.Delete([]byte(fullKey))
	})

	return err
}
