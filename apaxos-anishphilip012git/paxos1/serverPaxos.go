package main

import (
	"context"
	"fmt"
	"log"
	pb "paxos1/proto"
	"time"
)

func (s *PaxosServer) CalculateAcceptedTxns(txEntry map[int64]*pb.Transaction) map[int64]*pb.Transaction {
	// for k,v:=range txEntry{
	// 	if _,present:=s.
	// }
	return txEntry

}

var rpcCtxDeadlineTimeout time.Duration = 1 * time.Minute

func (s *PaxosServer) InitiatePaxosConsensus(index int64) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in consensus on %s: %v", s.serverID, r)
		}
	}()

	// Step 1: Check if the current server is already part of a higher ballot consensus
	if s.currentBallotNumber > s.acceptedBallotNumber {
		return false, fmt.Errorf("Server is already in a consensus round with a higher ballot %d > %d", s.currentBallotNumber, s.acceptedBallotNumber)
	}
	RandomDelay()

	// Step 2: Increment the current ballot number for this server
	s.currentBallotNumber++

	// Step 3: Prepare phase
	promiseTxnMap, promiseCount, err := s.PreparePhaseCall()
	if err != nil {
		log.Printf("Paxos Prepare failed on %s", s.serverID)
		return false, err
	}
	log.Printf("Paxos Prepare successful on %s with promise count %d", s.serverID, promiseCount)

	// Step 4: Calculate accepted transactions
	accpetedTxnMap := s.CalculateAcceptedTxns(promiseTxnMap)

	// Step 5: Accept phase
	acceptedCount, err := s.AcceptPhaseCall(accpetedTxnMap)
	if err != nil {
		log.Printf("Paxos Accept failed on %s with accpetCount %d", s.serverID, acceptedCount)
		return false, err
	}
	log.Printf("Paxos Accept successful on %s with accepted count %d", s.serverID, acceptedCount)

	// Step 6: Decide phase
	err = s.DecidePhaseCall(accpetedTxnMap)
	if err != nil {
		log.Printf("Paxos Decide failed on %s", s.serverID)
		return false, err
	}

	log.Printf("Paxos consensus succeeded on %s", s.serverID)
	return true, nil
}

func (s *PaxosServer) PreparePhaseCall() (map[int64]*pb.Transaction, int, error) {

	promiseCount := 0
	promiseTxnMap := map[int64]*pb.Transaction{}

	// Loop through all peers to send Prepare requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.serverID] {
			// For self, automatically count promise and gather uncommitted logs
			promiseCount++
			for k, v := range s.uncommittedLogs {
				promiseTxnMap[k] = v
			}
			for k, v := range s.acceptedValue {
				promiseTxnMap[k] = v
			}
			log.Println("Promise successful for self")
			continue
		}

		// Sending Prepare request to peers
		if LiveSvrMap[PeerIdMap[peerId]] {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel() // Ensure resources are freed when the operation completes

			resp, err := peer.Prepare(ctx, &pb.PrepareRequest{
				BallotNumber:        s.currentBallotNumber,
				LeaderId:            s.serverID,
				LastCommittedBallot: s.lastCommittedBallotNumber,
			})
			if !resp.Promise {
				log.Printf("Promise failed for %s : %d", s.serverID, s.currentBallotNumber)
				continue // Skip if the peer rejects the Prepare request
			}
			//Catch Up mechanism

			if err != nil && err.Error() == "SYNC" {
				retry := 3
				for retry > 0 {

					syncResp, _ := peer.SyncDataStore(ctx, &pb.TransactionRequest{
						TransactionEntry: s.committedLogs,
					})
					log.Printf("Server sync successful from %s to %s", s.serverID, PeerIdMap[peerId])
					retry--
					if syncResp.Success {
						break
					}
				}

			}
			for k, v := range resp.UncommittedTransactions {
				promiseTxnMap[k] = v
			}

			for k, v := range resp.LastAcceptedValue {
				promiseTxnMap[k] = v
			}
			promiseCount++
			log.Printf("Promise successful for %s : %d", s.serverID, s.currentBallotNumber)
		}
	}
	log.Printf("Promise count for %s : %d / %d -> %d", s.serverID, promiseCount, len(PeerAddresses), (len(PeerAddresses) / 2))
	// Check if majority promises were received
	if promiseCount <= len(PeerAddresses)/2 {
		return nil, promiseCount, fmt.Errorf("Prepare phase failed")
	}
	PaxosServerMapLock.Lock()
	for i, _ := range PaxosServerMap {
		if s.serverID == PaxosServerMap[i].serverID {
			PaxosServerMap[i].isLeader = true
		} else {
			PaxosServerMap[i].isLeader = false
		}
	}
	PaxosServerMapLock.Unlock()

	return promiseTxnMap, promiseCount, nil
}
func (s *PaxosServer) AcceptPhaseCall(accpetedTxnMap map[int64]*pb.Transaction) (int, error) {
	acceptedCount := 0

	// Loop through all peers to send Accept requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.serverID] {
			// For self, automatically count promise and gather uncommitted logs
			acceptedCount++
			s.acceptedValue = accpetedTxnMap
			s.acceptedBallotNumber = s.currentBallotNumber
			log.Println("Accept successful for self")
			continue
		}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
		defer cancel() // Ensure resources are freed when the operation completes
		resp, err := peer.Accept(ctx, &pb.AcceptRequest{
			BallotNumber:         s.currentBallotNumber,
			AccpetedTransactions: accpetedTxnMap,
		})
		if err != nil || !resp.Accepted {
			continue // Skip if peer fails to accept
		}
		acceptedCount++
	}

	// Check if majority acceptances were received
	if acceptedCount <= len(s.peerServers)/2 {
		return acceptedCount, fmt.Errorf("Accept phase failed")
	}

	return acceptedCount, nil
}
func (s *PaxosServer) DecidePhaseCall(accpetedTxnMap map[int64]*pb.Transaction) error {
	// Loop through all peers to send Decide requests
	for peerId, peer := range s.peerServers {
		// log.Printf("DecidePhaseCall [%s] [%s]", peerId, PeerAddresses[s.serverID])
		if peerId == PeerAddresses[s.serverID] {
			// For self, automatically count promise and gather uncommitted logs
			log.Printf("Committing transactions on server %v", s.serverID)
			s.CommitLogs(s.acceptedValue)
			log.Println("Decide successful for self")
			continue
		}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
		defer cancel() // Ensure resources are freed when the operation completes

		_, err := peer.Decide(ctx, &pb.DecisionRequest{
			TransactionsEntries: accpetedTxnMap,
		})
		if err != nil {
			return fmt.Errorf("Decide phase failed on peer %s", peer)
		}
	}
	return nil
}
