package security

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
)

// SignData signs any struct using the provided ECDSA private key and returns the signature.
func SignData(data interface{}, privateKey *ecdsa.PrivateKey) (string, error) {
	// Serialize the data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	// log.Println("json string::", string(dataBytes))

	// Hash the data using SHA-256
	hash := sha256.Sum256(dataBytes)

	// Generate the ECDSA signature
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}

	// Concatenate r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)

	// log.Println("SignData :: signature", string(signature))
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyData verifies the signature of any struct using the provided public key in PEM format.
func VerifyData(data interface{}, signatureStr string, pubKeyPEM []byte) (bool, error) {
	// Decode the signature from Base64
	// oldPrefix := log.Prefix()
	// log.SetPrefix(oldPrefix + "VerifyData :: ")
	// defer log.SetPrefix(oldPrefix)
	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		log.Printf("VerifyData error in decoding: %v", err)
		return false, err
	}
	// log.Println("signature", string(signature))
	// Parse PEM to obtain the public key block
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return false, errors.New("failed to decode PEM block containing public key")
	}

	// Parse the DER-encoded public key block
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse DER-encoded public key: %w", err)
	}

	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("invalid public key type, expected *ecdsa.PublicKey")
	}

	// Deserialize the signature into r and s
	signatureLen := len(signature)
	if signatureLen < 64 {
		return false, errors.New("invalid signature length")
	}
	r := big.Int{}
	s := big.Int{}
	r.SetBytes(signature[:signatureLen/2])
	s.SetBytes(signature[signatureLen/2:])

	// Serialize the data to JSON and hash it
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return false, err
	}
	// log.Println("json string::", string(dataBytes))
	hash := sha256.Sum256(dataBytes)

	// Verify the signature
	isValid := ecdsa.Verify(ecdsaPubKey, hash[:], &r, &s)
	// log.Println("isValid", isValid)
	return isValid, nil
}

// // ExampleStruct represents any struct you want to sign
// type ExampleStruct struct {
// 	Name  string
// 	Value int
// }

// func main() {
// 	// Create a sample struct to sign
// 	data := ExampleStruct{
// 		Name:  "Test",
// 		Value: 42,
// 	}

// 	// Generate an ECDSA key pair
// 	privateKey, pubKeyPEM, _ := GeneratePubPvtKeyPair("test")

// 	// Sign the data
// 	signature, err := SignData(data, privateKey)
// 	if err != nil {
// 		fmt.Println("Error signing data:", err)
// 		return
// 	}

// 	// Verify the signature
// 	isValid, err := VerifyData(data, signature, pubKeyPEM)
// 	if err != nil {
// 		fmt.Println("Error verifying data:", err)
// 		return
// 	}

// 	fmt.Println("Signature valid:", isValid)
// }

// // GeneratePubPvtKeyPair generates an ECDSA key pair and returns the private key and public key in PEM format.
// func GeneratePubPvtKeyPair(serverID string) (*ecdsa.PrivateKey, []byte, error) {
// 	// Generate a new ECDSA key pair
// 	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("error generating keys for server %s: %v", serverID, err)
// 	}

// 	// Serialize the public key to PEM format
// 	pubKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("error marshalling public key for server %s: %v", serverID, err)
// 	}
// 	pubKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER})
// 	return privateKey, pubKeyPEM, nil
// }
