package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// KeyPair holds the private and public keys of a server.
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte // Public key in PEM format
}

// GenerateServerKeys generates ECDSA key pairs for a list of servers and stores them.
func GeneratePubPvtKeyPair(serverID string) (*ecdsa.PrivateKey, []byte, error) {
	// Generate a new ECDSA key pair for each server
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating keys for server %s: %v", serverID, err)
	}

	// Serialize the public key to PEM format
	pubKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling public key for server %s: %v", serverID, err)
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER})
	return privateKey, pubKeyPEM, nil

}
