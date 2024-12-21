package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "paxos1/proto"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Utility function for logging and returning errors
func logErrorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	log.Println(err)
	return err
}

func defaultBalanceMap(serverIdList []string) map[string]int {
	svrMap := map[string]int{}
	for _, svr := range serverIdList {
		svrMap[svr] = 100
	}
	return svrMap
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
		metrics := GetMetricsInstance() // Get singleton instance

		start := time.Now()
		resp, err := handler(ctx, req) // Call the actual handler
		latency := time.Since(start)

		// Record the latency for this specific handler and server
		metrics.RecordCall(serverID, info.FullMethod, latency)

		return resp, err
	}
}

// // Start the gRPC server for Paxos
// func StartServer(serverID, address string, readyCh chan struct{}) {
// 	lis, err := net.Listen("tcp", address)
// 	if err != nil {
// 		log.Fatalf("Failed to listen: %v", err)
// 	}

// 	grpcServer := grpc.NewServer(
// 		grpc.UnaryInterceptor(ServerInterceptor(serverID)),
// 	)

// 	paxosSrv := &PaxosServer{serverID: serverID, addr: address, isDown: false, balance: 100}
// 	PaxosServerMapLock.Lock()
// 	PaxosServerMap[serverID] = paxosSrv
// 	PaxosServerMapLock.Unlock()
// 	// Register Paxos gRPC server
// 	pb.RegisterPaxosServer(grpcServer, paxosSrv)

// 	readyCh <- struct{}{}

// 	log.Printf("Paxos server %v is running on %v...", serverID, address)
// 	go func() {
// 		if err := grpcServer.Serve(lis); err != nil {
// 			log.Fatalf("Failed to serve: %v", err)
// 		}
// 	}()

// }

// PaxosServerWrapper wraps the gRPC server and its metadata.
type PaxosServerWrapper struct {
	serverID   string
	address    string
	grpcServer *grpc.Server
	listener   net.Listener
}

// StartServer starts the Paxos gRPC server and returns a wrapper to control it.
func StartServer(serverID, address string, readyCh chan struct{}) (*PaxosServerWrapper, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ServerInterceptor(serverID)),
	)

	// Initialize Paxos server
	paxosSrv := &PaxosServer{serverID: serverID, addr: address, isDown: false, balance: 100, balanceMap: defaultBalanceMap(AllServers)}

	// Register the Paxos gRPC service
	pb.RegisterPaxosServer(grpcServer, paxosSrv)

	// Store the Paxos server instance in the global map
	PaxosServerMapLock.Lock()
	PaxosServerMap[serverID] = paxosSrv
	PaxosServerMapLock.Unlock()

	readyCh <- struct{}{} // Signal that the server is ready.
	log.Printf("Paxos server %v is running on %v...", serverID, address)

	// Run the gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Return a wrapper that contains the server and its listener
	return &PaxosServerWrapper{
		serverID:   serverID,
		address:    address,
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

var PaxosServerWrapperMap = map[string]*PaxosServerWrapper{}

// StopServer gracefully stops the gRPC server.
func (s *PaxosServerWrapper) StopServer() {
	log.Printf("Shutting down Paxos server %v...", s.serverID)
	s.grpcServer.GracefulStop()
	s.listener.Close() // Close the listener to release the port.
	log.Printf("Paxos server %v stopped.", s.serverID)
}

// Run all servers and manage rounds separately
func StartAllServers(servers []string) {
	var wg sync.WaitGroup

	// Start all servers

	readyCh := make(chan struct{}, len(servers)) // Buffer to handle 5 ready signals
	for _, serverID := range servers {
		wg.Add(1) // Increment the WaitGroup counter
		go func(id string) {
			defer wg.Done()                                         // Decrement the counter when the goroutine completes
			psw, err := StartServer(id, PeerAddresses[id], readyCh) // Start the server
			if err != nil {
				log.Printf("StartServer %v", err)
			}
			PaxosServerWrapperMap[serverID] = psw
		}(serverID)
	}
	// Wait for all servers to send ready signals
	for i := 0; i < len(servers); i++ {
		<-readyCh // Wait for each server to signal readiness
	}
	wg.Wait() // Wait for all servers to finish starting
	log.Println("All servers have started.")
}

// Manage Paxos connections for a specific round
func MakePeerConnections(liveServers []string) {
	var livePeerAddresses []string
	for _, val := range liveServers {
		livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	}

	for _, server := range PaxosServerMap {
		server.ConnectToLivePeers(livePeerAddresses)
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
	for i := range PaxosServerMap {
		if LiveSvrMap[i] {
			PaxosServerMap[i].isDown = false
		} else {
			PaxosServerMap[i].isDown = true
		}
	}

}
func CloseRoundwiseConnections() {
	// var livePeerAddresses []string
	// for _, val := range liveServers {
	// 	livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	// }
	// server := &PaxosServer{}
	for _, server := range PaxosServerMap {
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
	for _, server := range PaxosServerMap {
		server.ConnectToLivePeers(livePeerAddresses) // Establish connections

		// log.Println("ser keys:", server.peerServers)
		for peerID, client := range server.peerServers {
			newPrefix := log.Prefix() + "On server " + server.serverID + ":: "
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
