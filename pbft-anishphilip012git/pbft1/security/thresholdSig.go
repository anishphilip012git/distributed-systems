package security

import (
	"fmt"
	"log"
	"os"

	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"
)

// TSReplica represents a decentralized node in the PBFT network
type TSReplica struct {
	Id          int
	SecretShare *ted25519.KeyShare
	Pub         ted25519.PublicKey
	Config      ted25519.ShareConfiguration
}

// InitializeTSReplica independently initializes a replica with its key share
func InitializeTSReplica(id int, config ted25519.ShareConfiguration, secretShare *ted25519.KeyShare, pub ted25519.PublicKey) *TSReplica {
	return &TSReplica{
		Id:          id,
		SecretShare: secretShare,
		Pub:         pub,
		Config:      config,
	}
}

type SharedNonce struct {
	AggregateNoncePub    ted25519.PublicKey
	AggregateNonceShares []*ted25519.NonceShare
}

// GenerateNonce generates a nonce for a specific message at the primary node
func GenerateNonces(config ted25519.ShareConfiguration, pub ted25519.PublicKey, message []byte, replicas []*TSReplica) (ted25519.PublicKey, []*ted25519.NonceShare, error) {
	var noncePubs []ted25519.PublicKey
	var nonceSharesCollection [][]*ted25519.NonceShare

	// Each replica generates its nonce
	for _, replica := range replicas {
		noncePub, nonceShares, _, err := ted25519.GenerateSharedNonce(&config, replica.SecretShare, pub, message)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate shared nonce for replica %d: %w", replica.Id, err)
		}
		noncePubs = append(noncePubs, noncePub)
		nonceSharesCollection = append(nonceSharesCollection, nonceShares)
	}

	// Aggregate nonce public keys
	aggregateNoncePub := noncePubs[0]
	for i := 1; i < len(noncePubs); i++ {
		aggregateNoncePub = ted25519.GeAdd(aggregateNoncePub, noncePubs[i])
	}

	// Aggregate nonce shares per replica
	aggregateNonceShares := make([]*ted25519.NonceShare, len(replicas))
	for j := 0; j < len(replicas); j++ {
		combinedShare := nonceSharesCollection[0][j]
		for i := 1; i < len(replicas); i++ {
			combinedShare = combinedShare.Add(nonceSharesCollection[i][j])
		}
		aggregateNonceShares[j] = combinedShare
	}

	return aggregateNoncePub, aggregateNonceShares, nil
}

// SignMessage creates a partial signature using the replica's own key and nonce
func (replica *TSReplica) SignMessage(message []byte, nonceShare *ted25519.NonceShare, noncePub ted25519.PublicKey) *ted25519.PartialSignature {
	return ted25519.TSign(message, replica.SecretShare, replica.Pub, nonceShare, noncePub)
}

// Primary aggregates partial signatures collected from other replicas
func AggregateSignatures(config ted25519.ShareConfiguration, partialSigs []*ted25519.PartialSignature) ([]byte, error) {
	sig, err := ted25519.Aggregate(partialSigs[:config.T], &config)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures: %w", err)
	}
	return sig, nil
}

// VerifySignature checks the validity of the aggregated signature
func VerifySignature(pub ted25519.PublicKey, message []byte, sig []byte) (bool, error) {
	ok, _ := ted25519.Verify(pub, message, sig)
	if ok {
		fmt.Println("Signature verified successfully")
		return ok, nil
	}
	return ok, fmt.Errorf("failed to verify signature")
}

var (
	TSReplicaConfig []*TSReplica
	// TsMu            sync.Mutex
)

func InitTSigConfig(N, T int) {
	config := ted25519.ShareConfiguration{T: T, N: N}
	log.Println("InitTSigConfig:: Initialising Threshold Sign Config")
	// Step 1: Generate shared key and secret shares for all replicas
	pub, secretShares, _, err := ted25519.GenerateSharedKey(&config)
	if err != nil {
		fmt.Printf("Error generating shared key: %v\n", err)
		return
	}
	// Step 2: Initialize each replica independently with its own secret share
	TSReplicaConfig = make([]*TSReplica, N)
	// TsMu = sync.Mutex{}
	for i := 0; i < N; i++ {

		TSReplicaConfig[i] = InitializeTSReplica(i, config, secretShares[i], pub)
	}
	log.Println("InitTSigConfig done of size", len(TSReplicaConfig))
}
func test() {
	msg := "Hello PBFT Threshold Signature"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}
	message := []byte(msg)

	// Define number of replicas (N) and threshold (T)
	N := 5
	T := 3
	config := ted25519.ShareConfiguration{T: T, N: N}

	// Step 1: Generate shared key and secret shares for all replicas
	pub, secretShares, _, err := ted25519.GenerateSharedKey(&config)
	if err != nil {
		fmt.Printf("Error generating shared key: %v\n", err)
		return
	}

	// Step 2: Initialize each replica independently with its own secret share
	replicas := make([]*TSReplica, N)
	for i := 0; i < N; i++ {
		replicas[i] = InitializeTSReplica(i, config, secretShares[i], pub)
	}

	// Step 3: Primary generates nonces for the replicas
	aggregateNoncePub, aggregateNonceShares, err := GenerateNonces(config, pub, message, replicas)
	if err != nil {
		fmt.Printf("Error generating nonces: %v\n", err)
		return
	}

	// Step 4: Each replica signs the message using the shared nonces
	partialSigs := make([]*ted25519.PartialSignature, 0, N)
	for _, replica := range replicas {
		partialSig := replica.SignMessage(message, aggregateNonceShares[replica.Id], aggregateNoncePub)
		partialSigs = append(partialSigs, partialSig)

		// Stop collecting once the threshold is reached
		if len(partialSigs) >= T {
			break
		}
	}
	// ted25519.PartialSignature.Bytes()

	// Step 5: Primary aggregates the collected partial signatures
	sig, err := AggregateSignatures(config, partialSigs)
	if err != nil {
		fmt.Printf("Error aggregating signatures: %v\n", err)
		return
	}

	// Step 6: Verify the final aggregated signature
	_, err = VerifySignature(pub, message, sig)
	if err != nil {
		fmt.Printf("Error in signature verification: %v\n", err)
	}
}
