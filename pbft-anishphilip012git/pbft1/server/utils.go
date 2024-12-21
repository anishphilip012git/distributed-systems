package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

// Define your structure
type MyStruct struct {
	Name string
	Age  int
}

// Function to hash any structure
func HashStruct(data interface{}) (string, error) {
	// Convert struct to JSON byte slice
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Create a new SHA256 hash
	hasher := sha256.New()
	// Write data to the hasher
	hasher.Write(jsonData)

	// Get the hashed result as a byte slice
	hashBytes := hasher.Sum(nil)
	// Convert the bytes to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

// RandomDelay adds a random delay between 1 and 50 milliseconds.
func RandomDelay() {
	// Seed the random number generator to avoid repetition.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random duration between 1 and 300 ms.
	delay := time.Duration(r.Intn(50-1+1)+1) * time.Millisecond

	log.Printf("Sleeping for %v...\n", delay)

	// Pause the program for the generated duration.
	time.Sleep(delay)
}
