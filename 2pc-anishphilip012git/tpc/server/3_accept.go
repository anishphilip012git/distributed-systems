package server

import (
	"context"
	"fmt"
	"sync"
	"time"
	pb "tpc/proto"
)

func (s *TPCServer) AcceptPhaseCall(accpetedTxnMap map[int64]*pb.Transaction) (int, error) {
	acceptedCount := 0
	var mu sync.Mutex // Mutex for synchronizing access to acceptedCount
	var wg sync.WaitGroup

	// Loop through all peers to send Accept requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			// For self, automatically count promise and gather uncommitted logs
			acceptedCount++
			s.acceptedValue = accpetedTxnMap
			s.acceptedBallotNumber = s.currentBallotNumber
			// log.Println("Accept successful for self")
			continue
		}

		wg.Add(1) // Add to the WaitGroup
		go func(peerId string, peer pb.TPCShardClient) {
			defer wg.Done() // Mark as done when the goroutine finishes
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel() // Ensure resources are freed when the operation completes

			resp, err := peer.PaxosAccept(ctx, &pb.AcceptRequest{
				BallotNumber:         s.currentBallotNumber,
				AcceptedTransactions: accpetedTxnMap,
			})
			if err != nil || !resp.Accepted {
				return // Skip if peer fails to accept
			}

			// Safely increment the accepted count
			mu.Lock()
			acceptedCount++
			mu.Unlock()
		}(peerId, peer)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if majority acceptances were received
	if acceptedCount <= s.clusterSize/2 {
		return acceptedCount, fmt.Errorf("Accept phase failed")
	}

	return acceptedCount, nil
}

// Accept phase: leader sends a list of transactions for consensus
func (s *TPCServer) PaxosAccept(ctx context.Context, req *pb.AcceptRequest) (*pb.AcceptedResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := &pb.AcceptedResponse{
		Accepted: true,
	}

	if req.BallotNumber < s.currentBallotNumber {
		response.Accepted = false
		return response, nil
	}

	s.acceptedBallotNumber = req.BallotNumber
	s.acceptedValue = req.AcceptedTransactions // Update the accepted transactions

	// log.Printf("Server %v accepted transactions with ballot %v", s.ServerID, req.BallotNumber)
	return response, nil
}
