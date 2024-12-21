package db

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func StoreKeyPair(db *badger.DB, privateKey, publicKey []byte) error {
	err := db.Update(func(txn *badger.Txn) error {
		// Store the private key
		if err := txn.Set([]byte("clientPrivateKey"), privateKey); err != nil {
			return err
		}
		// Store the public key
		return txn.Set([]byte("clientPublicKey"), publicKey)
	})
	return err
}
func RetrieveKeyPair(db *badger.DB) (*ecdsa.PrivateKey, []byte, error) {
	var privateKeyPEM, publicKeyPEM []byte

	// Retrieve the keys from Badger
	err := db.View(func(txn *badger.Txn) error {
		// Get the private key
		item, err := txn.Get([]byte("clientPrivateKey"))
		if err != nil {
			return err
		}
		privateKeyPEM, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Get the public key (in PEM format, to return directly)
		item, err = txn.Get([]byte("clientPublicKey"))
		if err != nil {
			return err
		}
		publicKeyPEM, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, nil, err
	}

	// Decode and parse the private key from PEM format
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, nil, fmt.Errorf("failed to decode private key PEM")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return privateKey, publicKeyPEM, nil
}
