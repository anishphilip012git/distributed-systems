package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"tpcbyz/db"
	"tpcbyz/metric"
	pb "tpcbyz/proto"
	. "tpcbyz/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var TPCServerMapLock = sync.RWMutex{}
var TPCServerMap = map[string]*TPCServer{}

// Utility function for logging and returning errors
func logErrorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	log.Println(err)
	return err
}

func defaultBalanceMap(clusterId int, n int) *ConcurrentMap {
	startIndex := (clusterId-1)*1000 + 1
	endIndex := startIndex + n - 1
	svrMap := map[int64]int{}
	for i := startIndex; i <= endIndex; i++ {
		svrMap[int64(i)] = 10
	}
	bmap := NewConcurrentMap()
	bmap.Data = svrMap
	return bmap
}

var skippedMethods = map[string]struct{}{
	// "/TPC.tpcbyz/SendTransaction": {},
	"/tpc.tpcbyz/CheckHealth": {},
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

// TPCServerWrapper wraps the gRPC server and its metadata.
type TPCServerWrapper struct {
	serverID   string
	address    string
	grpcServer *grpc.Server
	listener   net.Listener
}

func getServerID(serverName string, replicaCount int) (int, error) {
	// Remove 'S' from the server name
	numStr := strings.TrimPrefix(serverName, "S")

	// Convert the remaining string to an integer
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid server name format: %s", serverName)
	}

	// Perform modulo operation with ReplicaCount
	id := num % replicaCount

	return id, nil
}

// StartServer starts the TPC gRPC server and returns a wrapper to control it.
func StartServer(serverID, address string, noOfClients int, readyCh chan struct{}) (*TPCServerWrapper, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ServerInterceptor(serverID)),
	)
	wal, err := db.NewWALStore(db.DbConnections[serverID])
	if err != nil {
		log.Printf("Error in StartServer : NewWALStore", err)
	}
	// Initialize TPC server
	cSize := len(ClusterMapping[ServerClusters[serverID]].Servers)
	id, _ := getServerID(serverID, cSize)
	b_map := defaultBalanceMap(ServerClusters[serverID], noOfClients)
	TPCSrv := &TPCServer{ServerID: serverID, addr: address, BalanceMap: b_map, id: int64(id),
		db: db.DbConnections[serverID], WALStore: wal, ReplicaCount: int64(cSize), isDown: false,
		PubKeys: PublicKeys, PrivateKey: PrivateKeys[serverID],
		TxnState: make(map[int64]string), TxnStateBySeq: make(map[int64]string),
		CommittedTxns: make(map[int64]*pb.Transaction),
	}

	// Register the TPC gRPC service
	pb.RegisterTPCShardServer(grpcServer, TPCSrv)

	// Store the TPC server instance in the global map
	TPCServerMapLock.Lock()
	TPCServerMap[serverID] = TPCSrv
	TPCServerMapLock.Unlock()
	TxnIdMap = map[int64]*pb.Transaction{}
	readyCh <- struct{}{} // Signal that the server is ready.
	log.Printf("TPC server %v is running on %v...%v", serverID, address, len(b_map.Data))

	// Run the gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Return a wrapper that contains the server and its listener
	return &TPCServerWrapper{
		serverID:   serverID,
		address:    address,
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

var TPCServerWrapperMap = map[string]*TPCServerWrapper{}

// StopServer gracefully stops the gRPC server.
func (s *TPCServerWrapper) StopServer() {
	log.Printf("Shutting down TPC server %v...", s.serverID)
	s.grpcServer.GracefulStop()
	s.listener.Close() // Close the listener to release the port.
	log.Printf("TPC server %v stopped.", s.serverID)
}

// Run all servers and manage rounds separately
func StartAllServers(servers []string, noOfClients int) {
	var wg sync.WaitGroup

	// Start all servers

	readyCh := make(chan struct{}, len(servers)) // Buffer to handle 5 ready signals
	for _, serverID := range servers {
		wg.Add(1) // Increment the WaitGroup counter
		go func(id string) {
			defer wg.Done()                                                      // Decrement the counter when the goroutine completes
			psw, err := StartServer(id, PeerAddresses[id], noOfClients, readyCh) // Start the server
			if err != nil {
				log.Printf("StartServer %v", err)
			}
			TPCServerWrapperMap[serverID] = psw
		}(serverID)
	}
	// Wait for all servers to send ready signals
	for i := 0; i < len(servers); i++ {
		<-readyCh // Wait for each server to signal readiness
	}
	wg.Wait() // Wait for all servers to finish starting
	log.Println("All servers have started.")
}

// Manage TPC connections for a specific round
func MakePeerConnections(liveServers []string) {
	for serverID, server := range TPCServerMap {
		// Get the cluster ID for the current server
		serverClusterID := ServerClusters[serverID]

		// Collect addresses of live servers within the same cluster
		var livePeerAddresses []string
		for _, peerID := range liveServers {
			if ServerClusters[peerID] == serverClusterID {
				livePeerAddresses = append(livePeerAddresses, PeerAddresses[peerID])
			}
		}
		var liveclientSvrs []string
		for _, val := range ClientAddressMapSvrSide {
			liveclientSvrs = append(liveclientSvrs, val)
		}

		// Connect the current server to its live peers in the same cluster
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
	for i := range TPCServerMap {
		if LiveSvrMap[i] {
			TPCServerMap[i].isDown = false
		} else {
			TPCServerMap[i].isDown = true
		}
	}

}
func CloseRoundwiseConnections() {
	// var livePeerAddresses []string
	// for _, val := range liveServers {
	// 	livePeerAddresses = append(livePeerAddresses, PeerAddresses[val])
	// }
	// server := &TPCServer{}
	for _, server := range TPCServerMap {
		server.ClosePeerConnections()
	}
	// Assuming a decision is reached and the round is over
	// Close the connections for the round

}

// ValidatePeerConnections ensures that the connections are active and healthy.
func ValidatePeerConnections(liveServers []string) error {
	log.Println("ValidatePeerConnections", liveServers)
	for serverID, server := range TPCServerMap {
		// Get the cluster ID for the current server
		serverClusterID := ServerClusters[serverID]

		// Collect addresses of live servers within the same cluster
		var livePeerAddresses []string
		for _, peerID := range liveServers {
			// if ServerClusters[peerID] == serverClusterID {s
			livePeerAddresses = append(livePeerAddresses, PeerAddresses[peerID])
			// }
		}

		// Connect the current server to its live peers in the same cluster
		server.ConnectToLivePeers(livePeerAddresses)

		// Validate connections to peers in the same cluster
		for peerID, client := range server.peerServers {
			// if ServerClusters[peerID] != serverClusterID {
			// 	continue // Skip peers outside the cluster
			// }

			// Check the health of the connection
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			resp, err := client.CheckHealth(ctx, &pb.Empty{})
			if err != nil {
				if status.Code(err) == codes.Unavailable {
					log.Printf("Peer %s in cluster %d is unreachable: %v", peerID, serverClusterID, err)
					return fmt.Errorf("peer %s is down or not responding", peerID)
				}
				log.Printf("Error communicating with peer %s in cluster %d: %v", peerID, serverClusterID, err)
				return err
			}

			if !resp.Healthy {
				log.Printf("Peer %s in cluster %d is reporting as unhealthy", peerID, serverClusterID)
				return fmt.Errorf("peer %s is down internally", peerID)
			}
			// log.Printf("Peer %s in cluster %d is reporting as healthy", peerID, serverClusterID)
		}
	}

	return nil // All connections are healthy
}
func (s *TPCServer) ConnectToLiveClients(peerAddresses []string) {
	// s.clientServer = pb.PBFTClientServiceClient
	// s.clientConnection = *grpc.ClientConn{} // Store connections here
	newPrefix := "ConnectToLiveClients ::"
	// log.SetPrefix(newPrefix)

	for _, peerAddr := range peerAddresses {
		// // Create a context with a timeout for the connection
		// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// defer cancel()

		// Create a new gRPC connection

		conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf(newPrefix+"did not connect: %v", err)
		}

		// Store the connection so it can be closed later
		s.clientConnection = conn
		// log.Println("peer", peerAddr)
		// Create a new Paxos client from the connection
		client := pb.NewPBFTClientServiceClient(conn)
		s.clientServer = client
		// log.Printf("Server %v connected to peer server at %v", s.serverID, peerAddr)
	}
}
