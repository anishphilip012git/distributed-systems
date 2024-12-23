package client

import (
	"context"
	"log"
	"sync"
	"time"

	"tpc/data"
	"tpc/db"
	"tpc/metric"
	pb "tpc/proto"
	"tpc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	AllClients       []*Client
	ClientClusterMap []int // 0-indexed array  of
	//map[A:1 B:1 C:2 D:2 E:3 F:3]
	// ClientIndexMap   map[string]int // stores {A:1,B:2..liek this}
	// ClientAddressMap map[string]string
)

// func GenerateClientConfig(n, f int) {
// 	GenerateClientMap(n)
// 	// AllClients = make([]string, 0, n)

// 	AllClients = []*Client{}
// 	ClientAddressMap = map[string]string{}
// 	for i := 1; i <= n; i++ {
// 		clientID := fmt.Sprintf("Client %d", i)

//			address := fmt.Sprintf("localhost:600%02d", i)
//			ClientAddressMap[clientID] = address
//			AllClients = append(AllClients, NewClient(clientID, address, server.PeerAddresses, f, 10*time.Second, clientPvtKey))
//		}
//		// TimePerRequestPerServer = make(map[int64]map[string]time.Duration)
//	}

func GenerateClientConfig(n int, numServers int, numClusters int) {
	ClientClusterMap = DistributeClients(n, numClusters)

	for key, _ := range ClientClusterMap {
		if key == 0 {
			continue
		}
		AllClients = append(AllClients, NewClient(key))
	}
	// log.Println("ClientMap", len(ClientClusterMap))
	server.ShardMap = ClientClusterMap
}

type Client struct {
	id             int
	connMap        map[string]pb.TPCShardClient // gRPC client for communication
	bufferAtClient []*pb.Transaction            //if unable to buffer at server
}

// NewClient creates a new client and maps it to a specific server
func NewClient(id int) *Client {
	connMap := map[string]pb.TPCShardClient{}
	for svrId, addr := range server.PeerAddresses {
		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		c := pb.NewTPCShardClient(conn)
		connMap[svrId] = c
	}
	return &Client{
		id:             id,
		connMap:        connMap,
		bufferAtClient: []*pb.Transaction{},
	}
}

// SendTransaction sends a transaction to the client's assigned server via gRPC
func (c *Client) SendTransaction(index int64, tx data.Transaction) {
	txn := &pb.Transaction{
		Sender:   int64(tx.Sender),
		Receiver: int64(tx.Receiver),
		Amount:   int64(tx.Amount),
	}
	txnEntry := map[int64]*pb.Transaction{
		index: txn,
	}
	start := time.Now() // Record the start time
	// Send transaction via gRPC
	senderShard, rcvrShard := ClientClusterMap[tx.Sender], ClientClusterMap[tx.Receiver]
	if senderShard == rcvrShard {
		c.ProcessIntraShardTxn(senderShard, rcvrShard, txnEntry)
	} else {
		c.ProcessCrossShardTxn(senderShard, rcvrShard, txnEntry)
	}
	elapsed := time.Since(start) // Calculate the elapsed time
	metric.TnxSum.Mutex.Lock()
	metric.TnxSum.TotalElapsedTime += elapsed
	metric.TnxSum.TransactionCount++
	metric.TnxSum.Mutex.Unlock()
	db.StoreTime(db.DbConnections["S1"], index, txn, elapsed)

}

var clientTimeout = time.Second * 2

func (c *Client) ProcessCrossShardTxn(senderShard, rcvrShard int, txnEntry map[int64]*pb.Transaction) {
	log.Printf("[%d] Starting ProcessCrossShardTxn", c.id)
	request := &pb.TransactionRequest{
		TransactionEntry: txnEntry,
	}

	// Prepare Phase
	log.Printf("[%d] Starting Prepare Phase", c.id)
	prepareSuccess := c.preparePhase(senderShard, rcvrShard, request)

	if !prepareSuccess {
		log.Printf("[%d] Prepare Phase failed. Initiating Abort Phase.", c.id)
		c.abortPhase(senderShard, rcvrShard, request)
		return
	}

	// Commit Phase
	log.Printf("[%d] Prepare Phase succeeded. Initiating Commit Phase.", c.id)
	c.commitPhase(senderShard, rcvrShard, request)
}

func (c *Client) preparePhase(senderShard, rcvrShard int, request *pb.TransactionRequest) bool {
	prepareResults := make(map[int]chan bool)
	prepareResults[senderShard] = make(chan bool, 1)
	prepareResults[rcvrShard] = make(chan bool, 1)

	index := int64(0)
	for i := range request.TransactionEntry {
		index = i
	}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		prepareResults[senderShard] <- c.TPCPrepareToLeader(senderShard, request)
	}()
	go func() {
		defer wg.Done()
		prepareResults[rcvrShard] <- c.TPCPrepareToLeader(rcvrShard, request)
	}()

	wg.Wait()

	senderResult := <-prepareResults[senderShard]
	rcvrResult := <-prepareResults[rcvrShard]

	log.Printf("[%d] Prepare Phase Results: SenderShard=%v, ReceiverShard=%v", index, senderResult, rcvrResult)
	return senderResult && rcvrResult
}

func (c *Client) commitPhase(senderShard, rcvrShard int, request *pb.TransactionRequest) {
	var wg sync.WaitGroup
	responseChannels := make(map[string]chan *pb.TransactionResponse)
	index := int64(0)
	for i := range request.TransactionEntry {
		index = i
	}
	for svr := range c.connMap {
		if server.ServerClusters[svr] == senderShard || server.ServerClusters[svr] == rcvrShard {
			wg.Add(1)
			responseChannels[svr] = make(chan *pb.TransactionResponse, 1)

			go func(svr string, ch chan *pb.TransactionResponse) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
				defer cancel()

				resp, err := c.connMap[svr].TPCCommit(ctx, request)
				if err != nil {
					log.Printf("Client %d failed to send TPCCommit to server %s: %v", index, svr, err)
					return
				}
				ch <- resp
			}(svr, responseChannels[svr])
		}
	}

	wg.Wait()

	for svr, ch := range responseChannels {
		resp := <-ch
		log.Printf("[%d] Commit Phase Response from server %s: %v", index, svr, resp)
	}
}

func (c *Client) abortPhase(senderShard, rcvrShard int, request *pb.TransactionRequest) {
	var wg sync.WaitGroup
	responseChannels := make(map[string]chan *pb.TransactionResponse)
	index := int64(0)
	for i := range request.TransactionEntry {
		index = i
	}
	for svr := range c.connMap {
		if server.ServerClusters[svr] == senderShard || server.ServerClusters[svr] == rcvrShard {
			wg.Add(1)
			responseChannels[svr] = make(chan *pb.TransactionResponse, 1)

			go func(svr string, ch chan *pb.TransactionResponse) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
				defer cancel()

				resp, err := c.connMap[svr].TPCAbort(ctx, request)
				if err != nil {
					log.Printf("Client %d failed to send TPCAbort to server %s: %v", index, svr, err)
					return
				}
				ch <- resp
			}(svr, responseChannels[svr])
		}
	}

	wg.Wait()

	for svr, ch := range responseChannels {
		resp := <-ch
		log.Printf("[%d] Abort Phase Response from server %s: %v", index, svr, resp)
	}
}

func (c *Client) TPCPrepareToLeader(shard int, request *pb.TransactionRequest) bool {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	log.Println("TPCPrepareToLeader request", request)
	// Create the TransactionRequest object.
	index := int64(0)
	for i := range request.TransactionEntry {
		index = i
	}
	ldr := server.ClusterMapping[shard].Leader
	resp, err := c.connMap[ldr].TPCPrepare(ctx, request)

	if err != nil { //if err!=nil, then resp is nil in grpc
		log.Printf("TPCPrepareToLeader:: Client %d failed to send transaction %d to server %s: %v", index, c.id, ldr, err)
		return false
	}

	log.Printf("TPCPrepareToLeader :: Client %d received response from server: %v\n", c.id, resp)
	return true
}

func (c *Client) ProcessIntraShardTxn(senderShard, rcvrShard int, txnEntry map[int64]*pb.Transaction) bool {
	log.Printf("ProcessIntraShardTxn :: at client %d", c.id)
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	// Create the TransactionRequest object.
	request := &pb.TransactionRequest{
		TransactionEntry: txnEntry,
	}
	ldr := server.ClusterMapping[senderShard].Leader
	resp, err := c.connMap[ldr].SendTransaction(ctx, request)

	if err != nil { //|| !resp.Success
		log.Printf("Client %d failed to send transaction to server %s: %v", c.id, ldr, err)
		// Retry or handle buffering logic here
		// c.server.BufferTransaction(tx)
		return false
	}

	log.Printf("ProcessIntraShardTxn :: Client %d received response from server: %v\n", c.id, resp)
	return true
}
