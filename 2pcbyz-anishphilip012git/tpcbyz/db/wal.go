package db

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type WALEntry struct {
	TransactionID int64
	Sender        int64
	Receiver      int64
	Amount        int64
	Status        string //
	Timestamp     int64
}

const (
	PREPARE = "Prepared"
	COMMIT  = "Committed"
	ABORT   = "Aborted"
)

type WALStore struct {
	db       *bbolt.DB
	bucket   string
	mu       sync.RWMutex
	inMemory map[int64]*WALEntry // Optional in-memory cache for fast access
}

// NewWALStore initializes the WALStore with a BoltDB instance
func NewWALStore(db *bbolt.DB) (*WALStore, error) {
	bucketName := TBL_WAL
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WALStore: %w", err)
	}

	return &WALStore{
		db:       db,
		bucket:   bucketName,
		inMemory: make(map[int64]*WALEntry),
	}, nil
}

// AddEntry adds a new WAL entry and stores it persistently in BoltDB
func (w *WALStore) AddEntry(entry *WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(w.bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", w.bucket)
		}

		// Serialize the entry
		key := fmt.Sprintf("%d", entry.TransactionID)
		value := fmt.Sprintf("%d|%d|%d|%d|%s|%d",
			entry.TransactionID, entry.Sender, entry.Receiver, entry.Amount,
			entry.Status, entry.Timestamp) //Format(time.RFC3339)

		// Store in BoltDB
		return bucket.Put([]byte(key), []byte(value))
	})

	// Optionally, update in-memory cache
	if err == nil {
		w.inMemory[entry.TransactionID] = entry
	}

	return err
}

const TXN_NOT_FOUND = "TXN_NOT_FOUND"

// UpdateEntry updates the status of a WAL entry in BoltDB
func (w *WALStore) UpdateEntry(txID int64, newStatus, server string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(w.bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", w.bucket)
		}

		key := fmt.Sprintf("%d", txID)
		value := bucket.Get([]byte(key))
		if value == nil {
			log.Printf("in bucket size: %d transaction %d not found", bucket.Stats().KeyN, txID)
			return fmt.Errorf(TXN_NOT_FOUND)
		}

		// log.Printf("UpdateEntry:lastCommittedBallotNumber: Raw value for txn %d: %s", txID, string(value))

		// Manually parse the value
		fields := strings.Split(string(value), "|")
		if len(fields) != 6 {
			log.Printf("Invalid WAL entry for txn %d: %s (field count: %d)", txID, string(value), len(fields))
			return fmt.Errorf("invalid WAL entry format")
		}

		transactionID, _ := strconv.ParseInt(fields[0], 10, 64)
		sender, _ := strconv.ParseInt(fields[1], 10, 64)
		receiver, _ := strconv.ParseInt(fields[2], 10, 64)
		amount, _ := strconv.ParseInt(fields[3], 10, 64)
		status := fields[4]
		timestamp, _ := strconv.ParseInt(fields[5], 10, 64)

		// Update fields
		entry := WALEntry{
			TransactionID: transactionID,
			Sender:        sender,
			Receiver:      receiver,
			Amount:        amount,
			Status:        status,
			Timestamp:     timestamp,
		}
		entry.Status = newStatus
		entry.Timestamp = time.Now().UnixMilli()

		// Serialize updated entry
		newValue := fmt.Sprintf("%d|%d|%d|%d|%s|%d",
			entry.TransactionID, entry.Sender, entry.Receiver, entry.Amount,
			entry.Status, entry.Timestamp)

		log.Printf("UpdateEntry:: Updated value on %s for txn %d: %s", server, txID, newValue)
		return bucket.Put([]byte(key), []byte(newValue))
	})

	// Update in-memory cache if necessary
	if err == nil && w.inMemory[txID] != nil {
		w.inMemory[txID].Status = newStatus
		w.inMemory[txID].Timestamp = time.Now().UnixMilli()
	}

	return err
}

// ReplayWAL retrieves and processes all "Prepared" transactions
func (w *WALStore) ReplayWAL() error {
	return w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(w.bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", w.bucket)
		}

		return bucket.ForEach(func(k, v []byte) error {
			var entry WALEntry
			if _, err := fmt.Sscanf(string(v), "%d|%d|%d|%d|%s|%d",
				&entry.TransactionID, &entry.Sender, &entry.Receiver, &entry.Amount,
				&entry.Status, &entry.Timestamp); err != nil {
				return fmt.Errorf("failed to parse WAL entry: %w", err)
			}

			// Process only "Prepared" entries
			if entry.Status == "Prepared" {
				log.Printf("Replaying transaction: %v", entry)
				// Here, call RollbackOrCommit based on entry data
			}
			return nil
		})
	})
}

// RollbackOrCommit processes prepared transactions
func (w *WALStore) RollbackOrCommit(entry *WALEntry) {
	// Your rollback or commit logic here
	if entry.Status == "Prepared" {
		log.Printf("Rolling back transaction %d", entry.TransactionID)
		// Revert changes
	} else {
		log.Printf("Transaction %d already committed", entry.TransactionID)
	}
}

// Close WALStore and underlying BoltDB
func (w *WALStore) Close() error {
	return w.db.Close()
}

// CheckEntryStatus checks if a given transaction ID exists in the WAL with the specified status.
func (w *WALStore) CheckEntryStatus(txID int64, expectedStatus string) (bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	err := w.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(w.bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", w.bucket)
		}

		key := fmt.Sprintf("%d", txID)
		value := bucket.Get([]byte(key))
		if value == nil {
			return fmt.Errorf(TXN_NOT_FOUND)
		}

		// Manually parse the value
		fields := strings.Split(string(value), "|")
		if len(fields) != 6 {
			log.Printf("Invalid WAL entry for txn %d: %s (field count: %d)", txID, string(value), len(fields))
			return fmt.Errorf("invalid WAL entry format")
		}

		status := fields[4] // Extract the status field
		if status == expectedStatus {
			return nil // Status matches, no error
		}

		return fmt.Errorf("status mismatch: expected %s, found %s", expectedStatus, status)
	})

	if err != nil {
		if err.Error() == TXN_NOT_FOUND {
			return false, fmt.Errorf("transaction ID %d not found", txID)
		}
		if strings.Contains(err.Error(), "status mismatch") {
			return false, nil // Key exists but with a different status
		}
		return false, err // Other errors
	}

	return true, nil // Key exists and status matches
}
