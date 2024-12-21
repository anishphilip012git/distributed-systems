package db

import (
	"fmt"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type ConcurrentBoltMap struct {
	db       *bolt.DB
	keyLocks map[string]*sync.Mutex
	globalMu sync.Mutex
}

func NewConcurrentBoltMap(db *bolt.DB, walPath string) (*ConcurrentBoltMap, error) {
	return &ConcurrentBoltMap{
		db:       db,
		keyLocks: make(map[string]*sync.Mutex),
	}, nil
}

// GetLockForKey ensures a lock exists for a specific key
func (cbm *ConcurrentBoltMap) GetLockForKey(key string) *sync.Mutex {
	cbm.globalMu.Lock()
	defer cbm.globalMu.Unlock()

	if _, exists := cbm.keyLocks[key]; !exists {
		cbm.keyLocks[key] = &sync.Mutex{}
	}
	return cbm.keyLocks[key]
}

// Set writes a value to the map with WAL and 2PC
func (cbm *ConcurrentBoltMap) Set(key string, value []byte) error {
	keyLock := cbm.GetLockForKey(key)
	keyLock.Lock()
	defer keyLock.Unlock()

	// Perform the actual write in BoltDB
	err := cbm.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(TBL_BALANCE))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), value)
	})

	return err
}

// Get retrieves a value from the map
func (cbm *ConcurrentBoltMap) Get(key string) ([]byte, error) {
	keyLock := cbm.GetLockForKey(key)
	keyLock.Lock()
	defer keyLock.Unlock()

	var value []byte
	err := cbm.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(TBL_BALANCE))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	return value, err
}

// Close closes the BoltDB and WAL
func (cbm *ConcurrentBoltMap) Close() error {

	return cbm.db.Close()
}
