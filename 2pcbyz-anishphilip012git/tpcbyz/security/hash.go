package security

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

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
