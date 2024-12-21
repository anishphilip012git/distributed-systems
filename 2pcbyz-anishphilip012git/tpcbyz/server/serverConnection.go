package server

import (
	"context"
	"log"
	pb "tpcbyz/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *TPCServer) CheckHealth(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	if s.isDown {
		return &pb.HealthResponse{Healthy: false}, nil
	}
	return &pb.HealthResponse{Healthy: true}, nil
}

// Connect to peer servers and store the connection and clients
func (s *TPCServer) ConnectToLivePeers(peerAddresses []string) {
	s.peerServers = make(map[string]pb.TPCShardClient)
	s.peerConnections = make(map[string]*grpc.ClientConn) // Store connections here
	newPrefix := "ConnectToLivePeers ::"
	// log.Println(newPrefix, " on server", s.ServerID, "peer", peerAddresses)
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
		s.peerConnections[peerAddr] = conn
		// log.Println("peer", peerAddr)
		// Create a new Paxos client from the connection
		client := pb.NewTPCShardClient(conn)
		s.peerServers[peerAddr] = client
		// log.Printf("Server %v connected to peer server at %v", s.serverID, peerAddr)
	}
	// log.Println(newPrefix, " on server", s.ServerID, "len()", len(s.peerConnections))

}

// Close a specific connection by address
func (s *TPCServer) ClosePeerConnections() {
	for addr, conn := range s.peerConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection to server %v: %v", addr, err)
		} else {
			log.Printf("Closed connection to server %v from server %v", addr, s.ServerID)
		}
		delete(s.peerConnections, addr)
		delete(s.peerServers, addr)
	}
}
