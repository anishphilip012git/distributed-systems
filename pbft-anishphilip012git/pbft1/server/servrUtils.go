package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"pbft1/db"
	"pbft1/metric"
	pb "pbft1/proto"
	. "pbft1/security"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var PBFTSvrMapLock = sync.RWMutex{}
var PBFTSvrMap = map[string]*PBFTSvr{}

// Utility function for logging and returning errors
func logErrorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	log.Println(err)
	return err
}

func DefaultBalanceMap(n int) map[string]int {
	ClientMap := make(map[string]int)
	for i := 0; i < n; i++ {
		// Convert the integer to a character, starting from 'A'
		letter := string(rune('A' + i))
		ClientMap[letter] = 10
	}
	// log.Println("ClientMap", ClientMap)
	return ClientMap
}

var skippedMethods = map[string]struct{}{
	// "/paxos.Paxos/SendTransaction": {},
	"/paxos.Paxos/CheckHealth": {},
}

func ServerInterceptor(serverID string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, skip := skippedMethods[info.FullMethod]; skip {
			return handler(ctx, req) // Call the handler without tracking metrics.
		}
		metrics := metric.GetMetricsInstance() // Get singleton instance

		start := time.Now()
		resp, err := handler(ctx, req) // Call the actual handler
		latency := time.Since(start)

		// Record the latency for this specific handler and server
		metrics.RecordCall(serverID, info.FullMethod, latency)

		return resp, err
	}
}

// PaxosServerWrapper wraps the gRPC server and its metadata.
type PBFTServerWrapper struct {
	serverID   string
	address    string
	grpcServer *grpc.Server
	listener   net.Listener
}

// StartServer starts the Paxos gRPC server and returns a wrapper to control it.
func StartServer(i int64, clients int, serverID, address string, readyCh chan struct{}) (*PBFTServerWrapper, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ServerInterceptor(serverID)),
	)
	log.Println("i", i, len(TSReplicaConfig))
	// Initialize Paxos server
	// TsMu.Lock()
	log.Println(TSReplicaConfig[i])
	pbftSrv := &PBFTSvr{ServerID: serverID, Addr: address,
		Mu: sync.RWMutex{}, viewChangeMutex: sync.RWMutex{}, TxnStateMu: sync.RWMutex{},
		IsDown: false, id: i, BalanceMap: DefaultBalanceMap(clients), TSconfig: TSReplicaConfig[i], PrivateKey: PrivateKeys[serverID], PubKeys: PublicKeys,
		ClientPubKey: ClientKeys.PublicKey, db: db.DbConnections[serverID],
		ChkPointSize: 25,
		TxnState:     map[int64]string{}, waterMarkRange: 10,
		isBusy: false, busyLock: sync.RWMutex{},
	}
	// TsMu.Unlock()
	// Register the Paxos gRPC service
	pb.RegisterPBFTServiceServer(grpcServer, pbftSrv)

	// Store the Paxos server instance in the global map
	PBFTSvrMapLock.Lock()
	PBFTSvrMap[serverID] = pbftSrv
	PBFTSvrMapLock.Unlock()

	readyCh <- struct{}{} // Signal that the server is ready.
	log.Printf("PBFT server %v is running on %v...", serverID, address)

	// Run the gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Return a wrapper that contains the server and its listener
	return &PBFTServerWrapper{
		serverID:   serverID,
		address:    address,
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

// Run all servers and manage rounds separately
func StartAllServers(servers []string, clients int) {
	var wg sync.WaitGroup

	// Start all servers
	PBFTSvrMap = map[string]*PBFTSvr{}
	readyCh := make(chan struct{}, len(servers)) // Buffer to handle 5 ready signals
	for i, serverID := range servers {
		wg.Add(1) // Increment the WaitGroup counter
		go func(id string) {
			defer wg.Done()                                                            // Decrement the counter when the goroutine completes
			psw, err := StartServer(int64(i), clients, id, PeerAddresses[id], readyCh) // Start the server
			if err != nil {
				log.Printf("StartServer %v", err)
			}
			PBFTServerWrapperMap[serverID] = psw
		}(serverID)
	}
	// Wait for all servers to send ready signals
	for i := 0; i < len(servers); i++ {
		<-readyCh // Wait for each server to signal readiness
	}
	wg.Wait() // Wait for all servers to finish starting
	log.Println("All servers have started.")
}

func ResetServerStats() {
	for _, svrMap := range PBFTSvrMap {
		svrMap.Reset()
		log.Printf("ResetServerStats :: at %s %v ", svrMap.ServerID, svrMap.BalanceMap)
	}

}

// Manage Paxos connections for a specific round
func MakePeerConnections(liveServers []string) {
	var livePeerAddresses, liveclientSvrs []string
	for _, val := range liveServers {
		livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	}

	for _, val := range ClientAddressMapSvrSide {
		liveclientSvrs = append(liveclientSvrs, val)
	}

	for _, server := range PBFTSvrMap {
		server.ConnectToLivePeers(livePeerAddresses)
		server.ConnectToLiveClients(liveclientSvrs)
	}

}

// MakeRoundwiseConnection initializes a map of servers with true/false based on their live status.
func MakeRoundwiseConnection(liveServers []string) {
	// Create a set of live servers for quick lookups.
	liveSet := make(map[string]struct{})
	for _, server := range liveServers {
		// log.Println("MakeRoundwiseConnection ", server)
		liveSet[server] = struct{}{} // Use empty struct for minimal memory usage.
	}

	mu.Lock()
	defer mu.Unlock()
	// Initialize the status map for all servers.

	// Populate the status map.
	for _, server := range AllServers {
		_, isLive := liveSet[server] // Check if server exists in the live set.
		LiveSvrMap[server] = isLive
	}

	log.Println("Live Server map", LiveSvrMap)
	for i := range PBFTSvrMap {
		if LiveSvrMap[i] {
			PBFTSvrMap[i].IsDown = false
		} else {
			PBFTSvrMap[i].IsDown = true
		}
	}

}
func CloseRoundwiseConnections() {
	// var livePeerAddresses []string
	// for _, val := range liveServers {
	// 	livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	// }
	// server := &PaxosServer{}
	for _, server := range PBFTSvrMap {
		server.ClosePeerConnections()
	}
	// Assuming a decision is reached and the round is over
	// Close the connections for the round

}

// ValidatePeerConnections ensures that the connections are active and healthy.
func ValidatePeerConnections(liveServers []string) error {
	newPrefix := "ValidatePeerConnections ::"
	log.SetPrefix(newPrefix)
	var livePeerAddresses []string

	// Collect the addresses of live servers
	for _, val := range liveServers {
		livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	}
	// log.Println("livePeerAddresses", livePeerAddresses)

	// Connect to each peer server and validate the connection.
	for _, server := range PBFTSvrMap {
		server.ConnectToLivePeers(livePeerAddresses) // Establish connections

		// log.Println("ser keys:", server.peerServers)
		for peerID, client := range server.peerServers {
			newPrefix := log.Prefix() + "On server " + server.ServerID + ":: "
			log.SetPrefix(newPrefix)
			// Check the health of the connection
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Call CheckHealth to see if the peer server is reachable
			resp, err := client.CheckHealth(ctx, &pb.Empty{})

			if err != nil {
				// Handle gRPC-level errors (network unreachable, etc.)
				if status.Code(err) == codes.Unavailable {
					log.Printf("Peer %s is unreachable: %v", peerID, err)
					return fmt.Errorf("peer %s is down or not responding", peerID)
				}
				log.Printf("Error communicating with peer %s: %v", peerID, err)
				return err
			}

			if !resp.Healthy {
				log.Printf("Peer %s is reporting as unhealthy", peerID)
				return fmt.Errorf("peer %s is down internally", peerID)
			}

			// log.Printf("Peer %s is healthy and reachable", peerID)
		}
	}

	return nil // All connections are healthy
}
