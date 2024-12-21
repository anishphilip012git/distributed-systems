package pbft

import (
	"crypto/sha256"
	"fmt"
)

func (t *Transaction) Digest(useHash bool) string {
	// Create a colon-separated representation of the fields
	digestStr := fmt.Sprintf("%s:%s:%d:%d", t.Sender, t.Receiver, t.Amount, t.Index)

	// If hashing is enabled, hash the string for a fixed-size output
	if useHash {
		hash := sha256.Sum256([]byte(digestStr))
		return fmt.Sprintf("%x", hash[:16]) // Use the first 16 bytes for compactness
	}

	// Return the raw colon-separated string if hashing is not needed
	return digestStr
}

// Digest generates a colon-separated digest for the Transaction.
// Optionally hashes the result for a fixed-length digest.

// func (t *pb.Transaction) Digest() (string, error) {
// 	// Serialize the struct
// 	data, err := t.ToBytes()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Hash the serialized data
// 	hash := sha256.Sum256(data)
// 	truncatedHash := hash[:16] // Use the first 16 bytes for space efficiency

// 	// Return as a hex-encoded string
// 	return fmt.Sprintf("%x", truncatedHash), nil
// }
