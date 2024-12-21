package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tpcbyz/data"
	pb "tpcbyz/proto"
	"tpcbyz/security"
	"tpcbyz/server"
)

var clientTimeout = 2 * time.Second

// SendRequest sends a request to the current primary and waits for responses.
func (c *Client) SendRequest(index int64, tx data.Transaction) (string, error) {
	c.mutex.Lock()
	senderShard, recvrShard := ClientClusterMap[tx.Sender], ClientClusterMap[tx.Receiver]
	primaryIndex := server.ClusterMapping[senderShard].Leader //server.AllServers[c.viewNumber%int64(len(c.replicas))]
	log.Printf("Sending REQUEST to primary [%s] %d ,Shards: sender %d rcvr %d", primaryIndex, len(c.grpcClients), senderShard, recvrShard)
	// for k, _ := range c.grpcClients {
	// 	log.Printf("[%s]", k)
	// }
	primary := c.grpcClients[primaryIndex]

	c.timestamp++
	txnEntry := &pb.Transaction{
		Sender:   int64(tx.Sender),
		Receiver: int64(tx.Receiver),
		Amount:   int64(tx.Amount),
		Index:    index,
	}
	c.mutex.Unlock()
	// time.Sleep(1 * time.Minute)
	request := c.CreateClientRequest(txnEntry)
	// Channel to collect valid replies and error handling
	// responseCh := make(chan *pb.ClientReplyMessage, len(c.grpcClients))
	errorCh := make(chan error, len(c.grpcClients))

	// Send the request to the primary
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	start := time.Now()
	go func() {
		var (
			resp *pb.ClientReplyMessage
			err  error
		)
		// Primary processes the request and responds
		log.Printf(" Sending the client request  %s ", primaryIndex)
		resp, err = primary.ClientRequest(ctx, request)
		// if senderShard != recvrShard {
		// 	resp, err = primary.ClientRequest(ctx, request)
		// } else {
		// 	resp, err = primary.ClientRequest(ctx, request)
		// }

		if err != nil {
			errorCh <- fmt.Errorf("Error from primary replica: %v", err)
			return
		}
		if !c.VerifySvrmsg(primaryIndex, resp) {
			log.Println("VerifySvrmsg ::Failure")
			errorCh <- fmt.Errorf("Received invalid response from primary")
		}
		end := time.Since(start)
		AddTimeEntry(index, primaryIndex, end)
		log.Printf("ClientRequest ::  Time taken at client %s : %v", c.id, end.Seconds())

		// This should not broadcast to all replicas immediately
		// instead, wait for the actual Reply from the primary
		if resp != nil && resp.ViewNumber == c.viewNumber {
			c.responseCh <- resp // Send the valid primary response to the response channel
			log.Printf("Client Side:: %v  from primary", resp)

		} else {
			errorCh <- fmt.Errorf("Received invalid response from primary")
		}
	}()

	// Wait for valid replies to achieve quorum
	return c.waitForQuorum(errorCh, request, ctx)

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

func (c *Client) waitForQuorum(errorCh <-chan error, request *pb.ClientRequestMessage, ctx context.Context) (string, error) {
	// Determine the sender and receiver shards
	senderShard := ClientClusterMap[request.Transaction.Sender]
	receiverShard := ClientClusterMap[request.Transaction.Receiver]

	// Quorum requirement for each shard
	quorum := 2*c.f + 1

	// Track response counts for each shard
	responseCount := map[int]int{
		senderShard:   0,
		receiverShard: 0,
	}

	// Flags to check if quorum is reached for each shard
	shardQuorumReached := map[int]bool{
		senderShard:   false,
		receiverShard: false,
	}

	skipTxn := false

	for {
		select {
		case resp, ok := <-c.responseCh:
			if !ok {
				log.Println("Response channel closed unexpectedly")
				break
			}

			log.Printf("Client received response: %v", resp)

			// Handle "INSUFFICIENT_BAL" case
			if resp.Status == "INSUFFICIENT_BAL" {
				skipTxn = true
				log.Println("INSUFFICIENT_BAL :: Skipping transaction")
			}

			// Determine the shard of the responding node
			nodeShard := server.ServerClusters[resp.NodeId]
			if nodeShard != senderShard && nodeShard != receiverShard {
				log.Printf("Response from unrelated shard %d, ignoring", nodeShard)
				continue
			}

			// Increment response count for the appropriate shard
			responseCount[nodeShard]++

			// Check if quorum is reached for the shard
			if !shardQuorumReached[nodeShard] && responseCount[nodeShard] >= quorum {
				shardQuorumReached[nodeShard] = true
				log.Printf("Quorum reached for shard %d", nodeShard)
			}

		case err := <-errorCh:
			log.Printf("Error received: %v", err)
			// If there are errors, attempt to broadcast the request
			return c.broadcastRequest(request)

		case <-ctx.Done():
			// Timeout occurred
			log.Println("Context done, broadcasting request due to timeout")
			if !shardQuorumReached[senderShard] || !shardQuorumReached[receiverShard] {
				return c.broadcastRequest(request)
			}
		}

		// Exit loop if quorum is reached for both shards or if skipping the transaction
		if shardQuorumReached[senderShard] && shardQuorumReached[receiverShard] || skipTxn {
			break
		}
	}

	// Once quorum is reached for both shards, return success
	if shardQuorumReached[senderShard] && shardQuorumReached[receiverShard] {
		return "Success", nil
	}
	if skipTxn {
		return "Skipping transaction due to INSUFFICIENT_BAL", nil
	}

	// If quorum is not reached for one or both shards, return an error
	return "", fmt.Errorf("failed to reach quorum for both shards")
}

// // waitForQuorum waits for enough valid responses to reach a quorum.
// func (c *Client) waitForQuorum(errorCh <-chan error, request *pb.ClientRequestMessage, ctx context.Context) (string, error) {
// 	responseCount := make(map[int64]int)
// 	quorum := 2*c.f + 1    // Number of responses needed for quorum
// 	totalResponses := 0    // Total number of responses received
// 	quorumReached := false // Flag to track if quorum is reached
// 	skipTxn := false

// 	respondedNodes := make(map[string]bool)
// 	for {
// 		select {
// 		case resp, ok := <-c.responseCh:
// 			if !ok {
// 				log.Println("Response channel closed unexpectedly")
// 				break // Exit loop if channel is closed
// 			}
// 			log.Printf("Client Side:: %v from replicas", resp)
// 			if resp.Status == "INSUFFICIENT_BAL" {
// 				skipTxn = true
// 				log.Println("INSUFFICIENT_BAL ::Skipping request")
// 			}
// 			respondedNodes[resp.NodeId] = true

// 			responseCount[resp.RequestId]++
// 			totalResponses++
// 			log.Printf("Response Cnt at Client side %d %v ", totalResponses, responseCount)
// 			// Check if quorum is reached
// 			if !quorumReached && responseCount[resp.RequestId] >= quorum {
// 				quorumReached = true
// 				log.Println("Quorum reached at Client side ")

// 			}
// 		case err := <-errorCh:
// 			log.Printf("Error received: %v", err)
// 			// Trigger broadcasting the request if the primary is suspected faulty
// 			return c.broadcastRequest(request) // Implement this to handle the broadcast logic
// 		case <-ctx.Done():
// 			// Timeout occurred, consider broadcasting the request

// 			log.Println("Context done, broadcasting request due to timeout")
// 			if !quorumReached && !skipTxn {
// 				return c.broadcastRequest(request)
// 			}

// 		}

// 		// If all responses have been processed and the context hasn't timed out, exit
// 		if skipTxn || quorumReached {
// 			//time.Sleep(1 * time.Second)
// 			break
// 		}
// 	}

// 	// Once quorum is reached, return the status of the majority response
// 	defer log.Printf("RESPONSE :: for Txn %v %v", request.Transaction, respondedNodes)
// 	if quorumReached {
// 		return "Success", nil
// 	}
// 	if skipTxn {
// 		return "Skipping Txn due to INSUFFICIENT_BAL", nil
// 	}

// 	// In case no quorum was reached, handle accordingly (maybe timeout or retry)
// 	return "", fmt.Errorf("failed to reach quorum")
// }

// broadcastRequest sends the request to all replicas if the primary fails.
func (c *Client) broadcastRequest(request *pb.ClientRequestMessage) (string, error) {
	var wg sync.WaitGroup
	// responseCh := make(chan *pb.ClientReplyMessage, len(c.grpcClients))
	senderShard := ClientClusterMap[request.Transaction.Sender]
	errorCh := make(chan error, len(c.grpcClients))
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Broadcast the request to all replicas
	for i, client := range c.grpcClients {
		if client == nil || server.ServerClusters[i] != senderShard {
			continue
		}
		wg.Add(1)
		go func(s string, client pb.TPCShardClient) {
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
