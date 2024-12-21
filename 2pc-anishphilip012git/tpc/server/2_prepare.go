package server

import (
	"context"
	"fmt"
	"log"
	"time"
	pb "tpc/proto"
)

func (s *TPCServer) PreparePhaseCall() (map[int64]*pb.Transaction, int, error) {

	promiseCount := 0
	promiseTxnMap := map[int64]*pb.Transaction{}

	// Loop through all peers to send Prepare requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			// For self, automatically count promise and gather uncommitted logs
			promiseCount++
			for k, v := range s.uncommittedLogs {
				promiseTxnMap[k] = v
			}
			for k, v := range s.acceptedValue {
				promiseTxnMap[k] = v
			}
			// log.Println("Promise successful for self")
			continue
		}

		// Sending Prepare request to peers
		if LiveSvrMap[PeerIdMap[peerId]] {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel() // Ensure resources are freed when the operation completes

			resp, _ := peer.PaxosPrepare(ctx, &pb.PrepareRequest{
				BallotNumber:        s.currentBallotNumber,
				LeaderId:            s.ServerID,
				LastCommittedBallot: s.lastCommittedBallotNumber,
			})
			// log.Printf("PaxosPrepare resp : %v", resp)
			// log.Printf("PaxosPrepare resp on server %s : %v ", s.ServerID, resp)
			// if !resp.Promise {
			// 	log.Printf("Promise failed for %s : %d", s.ServerID, s.currentBallotNumber)
			// 	continue // Skip if the peer rejects the Prepare request
			// }

			//Catch Up mechanism
			if !resp.Promise { // err != nil && strings.Contains(err.Error(), "SYNC")
				// log.Printf("Committing transactions on server %v", s.ServerID)
				s.CommitLogs(resp.UncommittedTransactions)
				log.Printf("Server sync successful from %s to outdated leader =%s", PeerIdMap[peerId], s.ServerID)
			}
			for k, v := range resp.UncommittedTransactions {
				promiseTxnMap[k] = v
			}

			for k, v := range resp.LastAcceptedValue {
				promiseTxnMap[k] = v
			}
			promiseCount++
			// log.Printf("Promise successful for %s : %d  -> %v", s.ServerID, s.currentBallotNumber, promiseTxnMap)
		}
	}
	log.Printf("Promise count for %s : %d / %d -> %d", s.ServerID, promiseCount, s.clusterSize, (s.clusterSize / 2))
	// Check if majority promises were received
	if promiseCount <= s.clusterSize/2 {
		return nil, promiseCount, fmt.Errorf("Prepare phase failed")
	}
	// TPCServerMapLock.Lock()
	// for i, _ := range TPCServerMap {
	// 	if s.ServerID == TPCServerMap[i].ServerID {
	// 		TPCServerMap[i].isLeader = true
	// 	} else {
	// 		TPCServerMap[i].isLeader = false
	// 	}
	// }
	// TPCServerMapLock.Unlock()

	return promiseTxnMap, promiseCount, nil
}

// Prepare phase: leader proposes a ballot number
func (s *TPCServer) PaxosPrepare(ctx context.Context, req *pb.PrepareRequest) (*pb.PromiseResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := pb.PromiseResponse{
		Promise: true,
	}
	// log.Printf("SYNC Check::  %d <=%d , %d<%d", req.BallotNumber, s.currentBallotNumber, req.LastCommittedBallot, s.lastCommittedBallotNumber)
	// s.balanceMap[req.LeaderId] = int(req.LastSvrBalance) //Sync Balance Map
	if req.BallotNumber <= s.currentBallotNumber && req.LastCommittedBallot < s.lastCommittedBallotNumber { // to maintain datastore consistency ,when leader is outdated
		response.Promise = false
		// log.Printf("committedLogs %v", s.committedLogs)
		// log.Printf("UncommittedTransactions %v", response.UncommittedTransactions)
		if response.UncommittedTransactions == nil {
			response.UncommittedTransactions = map[int64]*pb.Transaction{}
		}
		for k, v := range s.committedLogs { //add both committed and acepted values
			response.UncommittedTransactions[k] = v
		}
		response.LastAcceptedValue = s.acceptedValue
		response.LastAcceptedBallot = s.acceptedBallotNumber
		msg := "LastCommittedBallot of leader is less than server.Need to SYNC"
		log.Printf("%s %v", msg, response)
		return &response, nil
	}
	s.currentBallotNumber = req.BallotNumber
	response.LastAcceptedBallot = s.acceptedBallotNumber
	response.LastAcceptedValue = s.acceptedValue // No need for accepted value as it's a transaction list
	response.UncommittedTransactions = s.uncommittedLogs

	// var err error
	// if req.LastCommittedBallot > s.lastCommittedBallotNumber {
	// 	err = fmt.Errorf("SYNC")
	// }else{

	// }

	// response.LastCommittedBallot = s.lastCommittedBallotNumber
	// log.Printf("Promise to Ldr %s  from %s for ballot %d gave %v", req.LeaderId, s.ServerID, req.BallotNumber, response)
	return &response, nil
}
