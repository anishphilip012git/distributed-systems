package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"net"
	"pbft1/data"
	pb "pbft1/proto"
	"pbft1/security"
	"sync"
	"time"

	. "pbft1/security"
	"pbft1/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	AllClients       []*Client
	ClientMap        map[string]int
	ClientAddressMap map[string]string
)

func GenerateClientConfig(n, f int) {
	GenerateClientMap(n)
	// AllClients = make([]string, 0, n)
	c_str := "Client"
	clientPvtKey, clientPUB, err := GeneratePubPvtKeyPair(c_str)
	if err != nil {
		log.Printf("Error in Key genration for server %s : %v", c_str, err)
	}
	server.ClientKeys = KeyPair{
		PrivateKey: clientPvtKey,
		PublicKey:  clientPUB,
	}
	AllClients = []*Client{}
	ClientAddressMap = map[string]string{}
	for i := 1; i <= n; i++ {
		clientID := fmt.Sprintf("Client %d", i)

		address := fmt.Sprintf("localhost:600%02d", i)
		ClientAddressMap[clientID] = address
		AllClients = append(AllClients, NewClient(clientID, address, server.PeerAddresses, f, 20*time.Second, clientPvtKey))
	}
	server.ClientAddressMapSvrSide = ClientAddressMap
	TimePerRequestPerServer = make(map[int64]map[string]time.Duration)
}
func GenerateClientMap(n int) {
	ClientMap = make(map[string]int)
	for i := 0; i < n; i++ {
		// Convert the integer to a character, starting from 'A'
		letter := string(rune('A' + i))
		ClientMap[letter] = i
	}
	log.Println("ClientMap", ClientMap)
}

type Client struct {
	pb.UnimplementedPBFTClientServiceServer
	id            string
	viewNumber    int64
	replicas      map[string]string
	grpcClients   map[string]pb.PBFTServiceClient
	timeout       time.Duration //
	timestamp     int64
	f             int    // Fault tolerance (maximum number of faulty replicas)
	primaryIndex  string // Index of the current primary based on view number
	mutex         sync.Mutex
	replyListener net.Listener                // Listener for incoming replies
	replyServer   *grpc.Server                // gRPC server to receive replica replies
	responseCh    chan *pb.ClientReplyMessage // Channel to collect replies
	clientAddress string
}

// Initialize the client with connection to all replicas
func NewClient(id, clientAddress string, replicaAddresses map[string]string, f int, timeout time.Duration, privateKey *ecdsa.PrivateKey) *Client {
	grpcClients := make(map[string]pb.PBFTServiceClient, len(replicaAddresses))
	for i, addr := range replicaAddresses {
		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		grpcClients[i] = pb.NewPBFTServiceClient(conn)
	}

	client := &Client{
		id:            id,
		viewNumber:    0, // Start with view 0
		replicas:      replicaAddresses,
		grpcClients:   grpcClients,
		timeout:       timeout, // 20*time.Second
		f:             f,
		timestamp:     0,
		responseCh:    make(chan *pb.ClientReplyMessage, len(grpcClients)),
		clientAddress: clientAddress,
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
	c.responseCh <- reply
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
func (c *Client) VerifySvrmsg(serverId string, msg *pb.ClientReplyMessage) bool {
	sign := msg.ReplicaSignature
	msg.ReplicaSignature = ""
	isValid, err := security.VerifyData(msg, sign, server.PublicKeys[serverId])
	if !isValid || err != nil {
		return false
	}
	return true
}

func (c *Client) SendRequest(index int64, tx data.Transaction) (string, error) {
	c.mutex.Lock()
	primaryIndex := server.AllServers[c.viewNumber%int64(len(c.replicas))]
	log.Printf("Sending REQUEST to primary %s ", primaryIndex)
	primary := c.grpcClients[primaryIndex]
	c.timestamp++
	txnEntry := &pb.Transaction{
		Sender:   tx.Sender,
		Receiver: tx.Receiver,
		Amount:   int64(tx.Amount),
		Index:    index,
	}
	c.mutex.Unlock()
	request := c.CreateClientRequest(txnEntry)

	// Channel to collect valid replies and error handling
	errorCh := make(chan error, len(c.grpcClients))

	// Send the request to the primary
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	start := time.Now()
	go func() {
		log.Printf("Sending client request to primary %s", primaryIndex)
		resp, err := primary.ClientRequest(ctx, request)
		if err != nil {
			errorCh <- fmt.Errorf("Error from primary replica: %v", err)
			return
		}

		if !c.VerifySvrmsg(primaryIndex, resp) {
			log.Println("VerifySvrmsg::Failure")
			errorCh <- fmt.Errorf("Received invalid response from primary")
		}

		end := time.Since(start)
		AddTimeEntry(index, primaryIndex, end)
		log.Printf("ClientRequest :: Time taken at client %s : %v", c.id, end.Seconds())

		if resp != nil && resp.ViewNumber == c.viewNumber {
			c.responseCh <- resp
			log.Printf("Client Side:: %v from primary", resp)
		} else {
			errorCh <- fmt.Errorf("Received invalid response from primary")
		}
	}()

	// Wait for valid replies to achieve quorum
	return c.waitForQuorum(errorCh, request, ctx)
}

func (c *Client) waitForQuorum(errorCh <-chan error, request *pb.ClientRequestMessage, ctx context.Context) (string, error) {
	var mu sync.Mutex
	responseCount := make(map[string]int)
	quorum := 2*c.f + 1
	totalResponses := 0
	quorumReached := false

	for {
		select {
		case resp := <-c.responseCh:
			mu.Lock()
			responseCount[resp.Status]++
			totalResponses++
			log.Printf("Client Side:: Response %v received, Response Count: %d", resp.Status, totalResponses)

			if !quorumReached && responseCount[resp.Status] >= quorum {
				quorumReached = true
				log.Println("Quorum reached on client side")
			}
			mu.Unlock()

		case err := <-errorCh:
			log.Printf("Error received: %v", err)
			return c.broadcastRequest(request) // Handle primary failure

		case <-ctx.Done():
			// if ctx.Err() != nil {
			// 	log.Println("Context cancelled due to timeout or external cancellation", ctx.Err())
			// }
			break // Exit loop if context is cancelled or times out
		}

		if quorumReached || totalResponses >= len(c.replicas) {
			break
		}
	}

	// Drain the response and error channels after quorum is reached
	go func() {
		for {
			select {
			case resp := <-c.responseCh:
				log.Printf("Drain response: %v", resp)
			case <-errorCh:
			case <-ctx.Done():
				return
			}
		}
	}()

	mu.Lock()
	defer mu.Unlock()
	if quorumReached {
		for status, count := range responseCount {
			if count >= quorum {
				return status, nil
			}
		}
	}

	// In case no quorum was reached, handle accordingly (maybe timeout or retry)
	return "", fmt.Errorf("failed to reach quorum")
}

// broadcastRequest sends the request to all replicas if the primary fails.
func (c *Client) broadcastRequest(request *pb.ClientRequestMessage) (string, error) {
	var wg sync.WaitGroup
	// responseCh := make(chan *pb.ClientReplyMessage, len(c.grpcClients))
	errorCh := make(chan error, len(c.grpcClients))
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Broadcast the request to all replicas
	for i, client := range c.grpcClients {
		if client == nil {
			continue
		}
		wg.Add(1)
		go func(s string, client pb.PBFTServiceClient) {
			defer wg.Done()
			// Send the request to the primary
			ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
			defer cancel()
			// Each replica responds with its last known valid reply
			resp, err := client.ClientRequest(ctx, request)
			if err != nil {
				errorCh <- fmt.Errorf("Error from replica %s: %v", s, err)
				return
			}

			if isValidReply(resp, request.Transaction.Index, c.viewNumber) {
				c.responseCh <- resp
			} else {
				errorCh <- fmt.Errorf("Invalid response from replica %s: %v", s, resp)
			}
		}(i, client)
	}

	// Wait for all goroutines to finish and close channels
	go func() {
		wg.Wait()
		// close(c.responseCh)
		close(errorCh)
	}()

	// Collect responses to reach quorum
	responseCount := make(map[string]int)
	quorum := c.f + 1

	for {
		select {
		case resp, ok := <-c.responseCh:
			if !ok {
				c.responseCh = nil
			} else {
				responseCount[resp.Status]++
				if responseCount[resp.Status] >= quorum {
					return resp.Status, nil
				}
			}
		case err, ok := <-errorCh:
			if !ok {
				errorCh = nil
			} else {
				fmt.Println("Error received:", err)
			}
		case <-ctx.Done():
			return "", fmt.Errorf("Timeout waiting for quorum responses")
		}

		if c.responseCh == nil && errorCh == nil {
			break
		}
	}

	return "", fmt.Errorf("Failed to reach quorum")
}

// isValidReply checks if the reply is valid according to PBFT protocol
func isValidReply(resp *pb.ClientReplyMessage, timestamp int64, viewNumber int64) bool {
	// Check for the current view number and timestamp validity
	return resp != nil && resp.ViewNumber == viewNumber && resp.RequestId == timestamp
}
