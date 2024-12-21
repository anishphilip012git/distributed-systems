package server

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	. "pbft1/security"
	"sync"
)

var (
	PeerAddresses = make(map[string]string)
	PeerIdMap     = make(map[string]string)
	AllServers    []string
	LiveSvrMap    = make(map[string]bool)
	ByzSvrMap     = make(map[string]bool)
	mu            sync.Mutex
)
var (
	ClientKeys KeyPair
)

func GenerateServerConfig(n int) {
	mu.Lock()
	defer mu.Unlock()
	ReplicaCount = int64(n)
	F = (n - 1) / 3

	// Clear existing maps and slices
	PeerAddresses = make(map[string]string)
	PeerIdMap = make(map[string]string)
	AllServers = make([]string, 0, n)
	LiveSvrMap = make(map[string]bool)
	ByzSvrMap = make(map[string]bool)
	for i := 1; i <= n; i++ {
		serverID := fmt.Sprintf("S%d", i)
		address := fmt.Sprintf("localhost:500%02d", i)

		// Populate the global variables
		PeerAddresses[serverID] = address
		PeerIdMap[address] = serverID
		AllServers = append(AllServers, serverID)
		LiveSvrMap[serverID] = true
		ByzSvrMap[serverID] = false
	}

	fmt.Println("PeerAddresses:", PeerAddresses)
	fmt.Println("PeerIdMap:", PeerIdMap)
	fmt.Println("AllServers:", AllServers)
	fmt.Println("LiveSvrMap:", LiveSvrMap)

	GenerateKeyPairsForAllServers()
	InitTSigConfig(n, F+1) //2*F+1
	MsgWiseNonceQ = map[int64]SharedNonce{}
	ChkpntMsgNonceQ = map[int64]SharedNonce{}
}

// ServerKeys holds the key pairs of the servers.
var PrivateKeys map[string]*ecdsa.PrivateKey // Maps server ID to its KeyPair
var PublicKeys map[string][]byte

func InitByzSvrMap(byzServer []string) {
	for serverId, _ := range LiveSvrMap {
		ByzSvrMap[serverId] = false
	}
	for _, svr := range byzServer {
		ByzSvrMap[svr] = true
	}
}

// Example usage
func GenerateKeyPairsForAllServers() {
	// List of server IDs
	// serverIDs := []string{"server1", "server2", "server3"}

	// Shared variable to store key pairs
	PublicKeys = map[string][]byte{}
	PrivateKeys = map[string]*ecdsa.PrivateKey{}

	// Generate keys
	for _, serverID := range AllServers {
		privateKey, pubKeyPEM, err := GeneratePubPvtKeyPair(serverID)
		if err != nil {
			log.Printf("Error in Key genration for server %s : %v", serverID, err)
		}
		// Store the key pair in the shared variable
		PrivateKeys[serverID] = privateKey // Store the private key
		PublicKeys[serverID] = pubKeyPEM
	}
}

var PBFTServerWrapperMap = map[string]*PBFTServerWrapper{}

// StopServer gracefully stops the gRPC server.
func (s *PBFTServerWrapper) StopServer() {
	log.Printf("Shutting down Paxos server %v...", s.serverID)
	s.grpcServer.GracefulStop()
	s.listener.Close() // Close the listener to release the port.
	log.Printf("PBFT server %v stopped.", s.serverID)
}
