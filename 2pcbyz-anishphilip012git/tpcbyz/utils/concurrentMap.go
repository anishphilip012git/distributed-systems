package utils

import (
	"errors"
	"log"
	"sync"
	"time"
)

// ReentrantLock: A lock that supports reentrant locking for the same process/thread.
type ReentrantLock struct {
	mutex    sync.Mutex
	owner    int64 // ID of the process/thread that owns the lock
	recCount int   // Reentrancy count
}

func NewReentrantLock() *ReentrantLock {
	return &ReentrantLock{}
}
func (r *ReentrantLock) IsLockTaken(ownerID int64) bool {

	// r.mutex.Lock()
	// defer r.mutex.Unlock()
	if r.owner == ownerID || r.owner == 0 {
		return false
	}
	log.Printf("LOCK_TAKEN:: attempted by %d, taken by %d", ownerID, r.owner)
	return true

}
func (r *ReentrantLock) Lock(ownerID int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// If the lock is already held by this owner, increment reentrant count
	if r.owner == ownerID {
		// log.Printf("LOCK:: held by same owner %d", ownerID)
		r.recCount++
		return
	}

	// Wait until the lock is released, then take ownership
	for r.recCount > 0 {
		r.mutex.Unlock()
		// log.Printf("LOCK IS BUSY for owner %d", ownerID)
		time.Sleep(10 * time.Millisecond) // Polling delay
		r.mutex.Lock()
	}
	r.owner = ownerID
	r.recCount = 1
	// log.Printf("LOCK:: r.recCount %d owner %d", r.recCount, r.owner)
}

func (r *ReentrantLock) Unlock(ownerID int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.owner != ownerID {
		log.Printf("UNLOCK:: called by a thread that does not own the lock :: %d != %d", r.owner, ownerID)
		return
	}

	r.recCount--
	if r.recCount == 0 {
		r.owner = 0 // Release ownership
		// r.recCount = 0
	}
	// log.Printf("UNLOCK:: for ownerId %d release ownership %d %d", ownerID, r.owner, r.recCount)
}

// ConcurrentMap: A map with fine-grained locking, ownership, and read/write capabilities.
type ConcurrentMap struct {
	Data     map[int64]int            // Key-value store
	Mu       sync.Mutex               // Global lock for map structure
	KeyLocks map[int64]*ReentrantLock // Locks for specific keys
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		Data:     make(map[int64]int),
		KeyLocks: make(map[int64]*ReentrantLock),
		Mu:       sync.Mutex{},
	}
}

// GetLockForKey ensures a reentrant lock exists for a specific key
func (cm *ConcurrentMap) GetLockForKey(key int64) *ReentrantLock {
	cm.Mu.Lock()
	defer cm.Mu.Unlock()

	if _, exists := cm.KeyLocks[key]; !exists {
		cm.KeyLocks[key] = NewReentrantLock()
	}
	return cm.KeyLocks[key]
}

const KEY_DNE = "KEY_DNE"
const LOCK_BUSY = "LOCK_BUSY"

// GetAndLock retrieves the value for a key and locks it for the given process/thread.
// The lock persists until UnlockKey is called.
func (cm *ConcurrentMap) GetAndLock(key int64, ownerID int64, server string) (int, error) {
	keyLock := cm.GetLockForKey(key)
	log.Printf("TRYING TO GET LOCK for %s key %d for tx %d", server, key, ownerID)
	if keyLock.IsLockTaken(ownerID) {
		log.Printf("WAITING TO GET LOCK for %s for tx %d: lock asked by %d, is currently held by %d", server, key, ownerID, keyLock.owner)
		// return -1, errors.New(LOCK_BUSY)// this means we are going to wait for next lock
	}
	keyLock.Lock(ownerID)
	cm.Mu.Lock()

	defer cm.Mu.Unlock()

	value, exists := cm.Data[key]
	if !exists {
		keyLock.Unlock(ownerID)
		return -1, errors.New(KEY_DNE)
	}
	log.Printf("SUCCESS :: TRYING TO GET LOCK for %s , key %d for tx %d", server, key, ownerID)
	return value, nil
}

// UnlockKey explicitly releases the lock for a key for the given process/thread.
func (cm *ConcurrentMap) UnlockKey(key int64, ownerID int64, server string) error {
	cm.Mu.Lock()
	defer cm.Mu.Unlock()

	lock, exists := cm.KeyLocks[key]
	if !exists {
		return errors.New(KEY_DNE)
	}

	lock.Unlock(ownerID)
	log.Printf("UnlockKey:: for txn %d for Server %s on client %d ", ownerID, server, key)
	return nil
}

// Set safely updates the value of a key. Requires the caller to lock the key first.
func (cm *ConcurrentMap) Set(key int64, value int, ownerID int64) (int, error) {
	keyLock := cm.GetLockForKey(key)
	// if !keyLock.IsLockTaken(ownerID) {
	keyLock.Lock(ownerID) // Reentrant safety
	defer keyLock.Unlock(ownerID)
	// }
	cm.Mu.Lock()
	defer cm.Mu.Unlock()

	cm.Data[key] = value
	return value, nil
}

// Read safely retrieves a value without locking (for shared reads).
func (cm *ConcurrentMap) Read(key int64) (int, bool) {
	cm.Mu.Lock()
	defer cm.Mu.Unlock()

	value, exists := cm.Data[key]
	if !exists {
		return 0, false //errors.New("key does not exist")
	}
	return value, true
}

// // Write safely updates a value for a key. Acquires a temporary lock.
// func (cm *ConcurrentMap) Write(key int64, value int) error {
// 	keyLock := cm.GetLockForKey(key)
// 	keyLock.Lock(0) // Use 0 for temporary locks
// 	defer keyLock.Unlock(0)

// 	cm.Mu.Lock()
// 	defer cm.Mu.Unlock()

// 	cm.Data[key] = value
// 	return nil
// }

// func main1() {
// 	cm := NewConcurrentMap()

// 	// Initialize some data
// 	cm.Set(1, 100, 1) // Set key=1 to value=100 with ownerID=1

// 	// Lock the key, modify its value, and release the lock
// 	ownerID := int64(1) // Process/thread ID
// 	value, err := cm.GetAndLock(1, ownerID)
// 	if err != nil {
// 		log.Fatalf("Error: %v", err)
// 	}

// 	// Modify the value while the key is locked
// 	fmt.Printf("Original Value: %d\n", value)
// 	newValue := value + 50
// 	cm.Set(1, newValue, ownerID)
// 	fmt.Printf("Updated Value: %d\n", newValue)

// 	// Unlock the key explicitly
// 	err = cm.UnlockKey(1, ownerID)
// 	if err != nil {
// 		log.Fatalf("Error unlocking key: %v", err)
// 	}

// 	// Read the updated value
// 	updatedValue, err := cm.Read(1)
// 	if err != nil {
// 		log.Fatalf("Error reading key: %v", err)
// 	}
// 	fmt.Printf("Final Value: %d\n", updatedValue)
// }
