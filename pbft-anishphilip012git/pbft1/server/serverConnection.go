package server

import (
	"context"
	"log"
	pb "pbft1/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *PBFTSvr) CheckHealth(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	if s.IsDown {
		return &pb.HealthResponse{Healthy: false}, nil
	}
	return &pb.HealthResponse{Healthy: true}, nil
}

// Connect to peer servers and store the connection and clients
func (s *PBFTSvr) ConnectToLivePeers(peerAddresses []string) {
	s.peerServers = make(map[string]pb.PBFTServiceClient)
	s.peerConnections = make(map[string]*grpc.ClientConn) // Store connections here
	newPrefix := "ConnectToLivePeers ::"
	log.SetPrefix(newPrefix)

	for _, peerAddr := range peerAddresses {
		// // Create a context with a timeout for the connection
		// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// defer cancel()

		// Create a new gRPC connection

		conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		// Store the connection so it can be closed later
		s.peerConnections[peerAddr] = conn
		// log.Println("peer", peerAddr)
		// Create a new Paxos client from the connection
		client := pb.NewPBFTServiceClient(conn)
		s.peerServers[peerAddr] = client
		// log.Printf("Server %v connected to peer server at %v", s.serverID, peerAddr)
	}
}

// Close a specific connection by address
func (s *PBFTSvr) ClosePeerConnections() {
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

// Connect to peer servers and store the connection and clients
func (s *PBFTSvr) ConnectToLiveClients(peerAddresses []string) {
	s.clientServers = make(map[string]pb.PBFTClientServiceClient)
	s.clientConnections = make(map[string]*grpc.ClientConn) // Store connections here
	newPrefix := "ConnectToLivePeers ::"
	log.SetPrefix(newPrefix)

	for _, peerAddr := range peerAddresses {
		// // Create a context with a timeout for the connection
		// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// defer cancel()

		// Create a new gRPC connection

		conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		// Store the connection so it can be closed later
		s.clientConnections[peerAddr] = conn
		// log.Println("peer", peerAddr)
		// Create a new Paxos client from the connection
		client := pb.NewPBFTClientServiceClient(conn)
		s.clientServers[peerAddr] = client
		// log.Printf("Server %v connected to peer server at %v", s.serverID, peerAddr)
	}
}
func (s *PBFTSvr) CloseClientConnections() {
	for addr, conn := range s.clientConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection to client %v: %v", addr, err)
		} else {
			log.Printf("Closed connection to client %v from server %v", addr, s.ServerID)
		}
		delete(s.clientConnections, addr)
		delete(s.clientServers, addr)
	}
}
