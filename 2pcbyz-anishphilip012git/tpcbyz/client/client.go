package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	pb "tpcbyz/proto"

	. "tpcbyz/security"
	"tpcbyz/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	AllClients       *Client
	ClientMap        map[string]int
	ClientAddressMap map[string]string
	ClientClusterMap []int
	responseChannels map[int]chan *pb.ClientReplyMessage // Key: shard ID
)

func GenerateClientConfig(n, numClusters, f int) {
	// GenerateClientMap(n)
	// AllClients = make([]string, 0, n)
	ClientClusterMap = DistributeClients(n, numClusters)
	c_str := "Client"
	clientPvtKey, clientPUB, err := GeneratePubPvtKeyPair(c_str)
	if err != nil {
		log.Printf("Error in Key genration for server %s : %v", c_str, err)
	}
	server.ShardMap = ClientClusterMap
	server.ClientKeys = KeyPair{
		PrivateKey: clientPvtKey,
		PublicKey:  clientPUB,
	}
	// AllClients = *Client{}
	ClientAddressMap = map[string]string{}
	// for i := 1; i <= n; i++ {
	i := 1
	clientID := fmt.Sprintf("Client %d", i)

	address := fmt.Sprintf("localhost:600%02d", i)
	ClientAddressMap[clientID] = address
	// AllClients = append(AllClients, NewClient(clientID, address, server.PeerAddresses, f, 10*time.Second, clientPvtKey))
	AllClients = NewClient(clientID, address, server.PeerAddresses, f, clientTimeout, clientPvtKey, numClusters)
	// }
	server.ClientAddressMapSvrSide = ClientAddressMap
	TimePerRequestPerServer = make(map[int64]map[string]time.Duration)

	responseChannels := make(map[int]chan *pb.ClientReplyMessage)
	for shardID := 0; shardID < numClusters; shardID++ {
		responseChannels[shardID] = make(chan *pb.ClientReplyMessage, 2*f+1)
	}
}

type Client struct {
	pb.UnimplementedPBFTClientServiceServer
	id          string
	viewNumber  int64
	replicas    map[string]string
	numClusters int
	grpcClients map[string]pb.TPCShardClient
	timeout     time.Duration //
	timestamp   int64
	f           int // Fault tolerance (maximum number of faulty replicas)
	// primaryIndex  string // Index of the current primary based on view number
	mutex            sync.Mutex
	replyListener    net.Listener                // Listener for incoming replies
	replyServer      *grpc.Server                // gRPC server to receive replica replies
	responseCh       chan *pb.ClientReplyMessage // Channel to collect replies
	clientAddress    string
	responseChannels map[int]chan *pb.ClientReplyMessage // Key: shard ID
}

// Initialize the client with connection to all replicas
func NewClient(id, clientAddress string, replicaAddresses map[string]string, f int, timeout time.Duration, privateKey *ecdsa.PrivateKey, numClusters int) *Client {
	grpcClients := make(map[string]pb.TPCShardClient, len(replicaAddresses))
	for i, addr := range replicaAddresses {
		log.Println("NewClient", i, addr)
		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		grpcClients[i] = pb.NewTPCShardClient(conn)
	}
	// Initialize response channels for each shard

	client := &Client{
		id:          id,
		numClusters: numClusters,
		viewNumber:  0, // Start with view 0
		replicas:    replicaAddresses,
		grpcClients: grpcClients,
		timeout:     timeout, // 40*time.Second
		f:           f,
		timestamp:   0,
		// responseCh:    make(chan *pb.ClientReplyMessage, 2*f+1),
		responseChannels: responseChannels,
		clientAddress:    clientAddress,
		// primaryIndex: "S1", // Primary is initially the first replica
	}

	err := client.StartListener(client.clientAddress) // Change port as needed
	if err != nil {
		log.Printf("failed to start client listener: %v", err)
	}

	return client
}
func (c *Client) StartListener(port string) error {
	// Set up listener on specified port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}
	c.replyListener = listener
	c.replyServer = grpc.NewServer()

	// Register the service handler for handling replies
	pb.RegisterPBFTClientServiceServer(c.replyServer, c) // Assuming client implements PBFTServiceServer

	// Start serving in a separate goroutine
	go func() {
		if err := c.replyServer.Serve(c.replyListener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return nil
}

// Implement the minimal PBFTClientServiceServer interface
func (c *Client) ReplyMessage(ctx context.Context, reply *pb.ClientReplyMessage) (*pb.StatusResponse, error) {
	// Send the reply to the response channel for quorum processing
	log.Printf("ReplyMessage :: received reply at client %v ", reply)
	shardID := server.ServerClusters[reply.NodeId]
	// c.responseCh <- reply
	if ch, exists := c.responseChannels[shardID]; exists {
		ch <- reply
	} else {
		log.Printf("ReplyMessage :: no channel found for shard %d", shardID)
	}
	return &pb.StatusResponse{Success: true}, nil
}
func (c *Client) CreateClientRequest(txnEntry *pb.Transaction) *pb.ClientRequestMessage {
	request := &pb.ClientRequestMessage{
		ClientId:        c.id,
		Timestamp:       c.timestamp,
		Transaction:     txnEntry,
		ClientSignature: "",
	}
	sign, _ := SignData(request, server.ClientKeys.PrivateKey)
	request.ClientSignature = sign
	return request
}

var TimePerRequestPerServer map[int64]map[string]time.Duration

func AddTimeEntry(txnID int64, server string, duration time.Duration) {
	// Initialize the nested map for txnID if it doesn't exist
	if _, exists := TimePerRequestPerServer[txnID]; !exists {
		TimePerRequestPerServer[txnID] = make(map[string]time.Duration)
	}

	// Add or update the duration for the specified server
	TimePerRequestPerServer[txnID][server] = duration
}
