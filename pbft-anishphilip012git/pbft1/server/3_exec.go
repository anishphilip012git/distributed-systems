package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"pbft1/db"
	pb "pbft1/proto"
	. "pbft1/security"
	"sync"
	"time"
)

// CommitPhaseCall: Executes the request and sends a reply to the client
func (s *PBFTSvr) CommitPhaseCall(clientRequest *pb.ClientRequestMessage, seqNo int64) (*pb.StatusResponse, error) {
	log.Println("CommitPhaseCall ::")

	execCnt := 0
	executionOrder := &pb.CommitMessage{
		ViewNumber:     s.ViewNumber,
		SequenceNumber: seqNo,
		CommitCertificate: &pb.CommitCertificate{
			ViewNumber:         s.ViewNumber,
			PrepareReqMessages: s.CommitedLogs[seqNo], // Finalized commit certificate
		},
		Request: clientRequest,
	}

	var wg sync.WaitGroup // To wait for all goroutines to finish
	mutex := sync.Mutex{} // Mutex to safely update execCnt

	// Loop through all peers to send Commit requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			s.Mu.Lock()
			// For self, automatically count the commit and log it
			s.TxnState[clientRequest.Transaction.Index] = C_STATE
			s.TxnStateBySeq[seqNo] = C_STATE
			execCnt++ // Count commit for self

			s.Mu.Unlock()

			continue
		}

		// Spawning a goroutine for each peer to send the commit request
		if LiveSvrMap[PeerIdMap[peerId]] && !ByzSvrMap[PeerIdMap[peerId]] {
			wg.Add(1)
			go func(peer pb.PBFTServiceClient, peerId string) {
				defer wg.Done() // Ensure the WaitGroup counter is decremented when done

				// Create context with deadline
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
				defer cancel()

				// Send Commit request to the peer
				resp, err := peer.CommitRequest(ctx, executionOrder)
				if err != nil {
					log.Printf("Commit failed for %s : %v", s.ServerID, err)
					return // Skip if the peer rejects the Commit request
				}

				// If the response is successful, increment the commit count
				if resp.Success {
					mutex.Lock() // Lock to safely update execCnt
					execCnt++
					mutex.Unlock()
					log.Printf("Commit successful for %s :: response- %v", s.ServerID, resp)
				}
			}(peer, peerId) // Pass peer and peerId to the goroutine
		}
	}

	// Wait for all goroutines to finish before proceeding
	wg.Wait()

	// After all commits are processed, check if the execution was successful
	statusResponse := &pb.StatusResponse{Success: execCnt >= (2*F + 1)} // Quorum reached if > len/2

	if execCnt < (2*F + 1) {
		return statusResponse, fmt.Errorf("Commit failed: Quorum not reached")
	}
	executed := s.ExecuteTxn(clientRequest.Transaction, seqNo)
	log.Println("Commit successful for self", executed)
	// Send ReplyMessage back to client (this would typically be a callback or message queue)
	log.Printf("Commit successful for Transaction ID %d at server %s", clientRequest.Transaction.Index, s.ServerID)
	return statusResponse, nil
}

func (s *PBFTSvr) ChkTxnIdExecuted(idx int64) bool {
	chk1, ok := s.TxnState[idx]
	if !(ok && chk1 == E_STATE) {
		return false
	}
	val, err := db.GetRecord(s.db, db.TBL_EXEC, idx)
	if err != nil {
		return false
	}
	if val == "" {
		return false
	}
	return true
}

// ExecuteTxn executes a transaction if it has not already been executed
func (s *PBFTSvr) ExecuteTxn(txn *pb.Transaction, seqNo int64) bool {

	idx := txn.Index
	start := time.Now()
	// Check if transaction is already executed
	if s.ChkTxnIdExecuted(idx) {
		log.Printf("Txn %d: %v already executed", idx, txn)
		return true
	}

	// Execute the transaction if it's the first or the previous transaction is executed
	// if idx == 1 || s.ChkTxnIdExecuted(idx-1) {
	s.markTransactionExecuted(txn, seqNo)
	log.Printf("ExecuteTxn :: time to execute at %s -txn %d::%v", s.ServerID, txn.Index, time.Since(start))
	if s.TxnState[idx+1] == C_STATE {
		nextTxn := TxnIdMap[idx+1] // assuming you have a way to retrieve transactions by index

		if nextTxn != nil {
			// Execute the next transaction
			log.Printf("ExecuteTxn :: Executing next transaction %d: %v", nextTxn.Index, nextTxn)
			s.markTransactionExecuted(nextTxn, TxnIdMapWithSeqNo[idx])
		}
	}
	if s.isCheckpointNeeded(seqNo) {
		go s.CreateCheckpoint(seqNo)
	}
	return true
	// }

	// log.Printf("Txn %d: %v not executed: previous transaction not executed", idx, txn)
	// return false
}

// markTransactionExecuted updates the execution state of a transaction and logs it
func (s *PBFTSvr) markTransactionExecuted(txn *pb.Transaction, seqNo int64) {

	idx := txn.Index
	s.Mu.Lock()
	if s.BalanceMap[txn.Sender] > int(txn.Amount) {
		s.BalanceMap[txn.Sender] -= int(txn.Amount)
		s.BalanceMap[txn.Receiver] += int(txn.Amount)
	}
	s.TxnState[idx] = E_STATE
	s.TxnStateBySeq[seqNo] = E_STATE
	s.Mu.Unlock()
	db.StoreRecord(s.db, db.TBL_EXEC, txn.Index, txn)
	log.Printf("Txn %d: %v executed at server %s", idx, txn, s.ServerID)
}

// isCheckpointNeeded checks if a checkpoint should be created
func (s *PBFTSvr) isCheckpointNeeded(seqNo int64) bool {
	return seqNo%s.ChkPointSize == 0
}

// ExecuteRequest handles the execution of the client request after consensus is reached
func (s *PBFTSvr) CommitRequest(ctx context.Context, in *pb.CommitMessage) (*pb.StatusResponse, error) {
	// 1. Check if the CommitMessage is valid
	if in == nil {
		return nil, errors.New("invalid CommitMessage, cannot be nil")
	}
	seqNo := in.SequenceNumber
	// Verify that the CommitMessage has the correct view_number and sequence_number
	if !s.validateCommitMessage(in) {
		log.Printf("Invalid CommitMessage: View %d, Sequence %d", in.ViewNumber, in.SequenceNumber)
		return &pb.StatusResponse{Success: false}, fmt.Errorf("execution order validation failed")
	}

	// 2. Validate the commit certificate
	if !s.validateCommitCertificate(in.CommitCertificate) {
		log.Println("Invalid commit certificate")
		return &pb.StatusResponse{Success: false}, fmt.Errorf("commit certificate validation failed")
	}
	idx := in.CommitCertificate.PrepareReqMessages.Request.Transaction.Index

	// 3. Apply the transaction to the replica's state
	// You might use the `in.CommitCertificate` or `in.CommitMessage` to fetch the actual transaction details.
	transaction := in.CommitCertificate.PrepareReqMessages.Request.Transaction // Assuming we're using the first PrepareMessage's transaction

	s.Mu.Lock()
	s.TxnState[idx] = C_STATE
	s.TxnStateBySeq[seqNo] = C_STATE
	s.Mu.Unlock()
	executed := s.ExecuteTxn(transaction, seqNo)
	s.resetViewChangeTimer()

	// if err != nil {
	// 	log.Printf("Failed to apply transaction: %v", err)
	// 	return &pb.StatusResponse{Success: false}, err
	// }
	// Reply phase: Replica sends reply to client
	if s.id != s.getPrimaryIdForView(in.ViewNumber) {
		log.Printf("CreateClientReplyMessage")
		reply := s.CreateClientReplyMessage(s.TxnState[idx], in.ViewNumber, idx)

		go s.SendReplyToClient(ClientAddressMapSvrSide[in.Request.ClientId], in.Request.Transaction.Index, reply)
	}
	// 4. Return success
	log.Printf("CommitRequest successful transaction with sequence number %d on Server %s", in.SequenceNumber, s.ServerID)
	return &pb.StatusResponse{Success: executed}, nil
}
func (s *PBFTSvr) SendReplyToClient(clientID string, txnIdx int64, reply *pb.ClientReplyMessage) {
	// s.Mu.Lock()
	// defer s.Mu.Unlock()
	start := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeoutForClient))
	defer cancel() // Ensure resources are freed when the operation completes

	// log.Printf("PrePrepareRequest Promise for %s debug 0", s.ServerID)
	resp, err := s.clientServers[clientID].ReplyMessage(ctx, reply)
	if err != nil {
		log.Printf("SendReplyToClient failed for %s txn %d: %v", s.ServerID, txnIdx, err)
		log.Printf("Time taken in failure at %s : %v", s.ServerID, time.Now().Sub(start).Seconds())
		return
	}
	log.Printf("SendReplyToClient ::  Time taken at %s : %v", s.ServerID, time.Now().Sub(start).Seconds())
	log.Printf("SendReplyToClient :: passed for server %s with %v", s.ServerID, resp)
}
func (s *PBFTSvr) CreateClientReplyMessage(msg string, viewNumber, idx int64) *pb.ClientReplyMessage {
	// digest, _ := HashStruct(msg.Request)
	reply := &pb.ClientReplyMessage{
		ViewNumber:       viewNumber,
		Status:           msg,
		RequestId:        idx,
		ReplicaSignature: "",
	}
	sign, _ := SignData(reply, s.PrivateKey)
	reply.ReplicaSignature = sign
	return reply
}

// validateCommitMessage verifies that the CommitMessage has the correct view and sequence numbers
func (s *PBFTSvr) validateCommitMessage(in *pb.CommitMessage) bool {
	// Here we just do a simple check for the view number and sequence number.
	// You may need to add more logic to handle the view transition and sequence ordering
	if in.ViewNumber < 0 || in.SequenceNumber <= 0 {
		return false
	}
	// Additional checks can be implemented here depending on your PBFT state machine

	return true
}

// validateCommitCertificate checks the validity of the commit certificate
func (s *PBFTSvr) validateCommitCertificate(cert *pb.CommitCertificate) bool {
	// For simplicity, we check if the commit messages are not empty and if there are enough commit messages
	// In a real-world scenario, you would validate the signatures, quorum size, etc.

	// Ensure the commit certificate has a valid quorum (this is simplified for illustration)
	log.Printf("validateCommitCertificate ::%d %d ", len(cert.PrepareReqMessages.PrepareCertificate.PrepareMessages), cert.PrepareReqMessages.PrepareCertificate.PrepareMessages[0].SequenceNumber)
	if len(cert.PrepareReqMessages.PrepareCertificate.PrepareMessages) <= 2*F {
		return false // Assuming quorum requires at least 2 commit messages
	}

	digest1, _ := HashStruct(cert.PrepareReqMessages.Request)
	cnt := 0
	for _, val := range cert.PrepareReqMessages.PrepareCertificate.PrepareMessages {
		if digest1 == val.ClientRequestHash {

			cnt++
		}
	}
	log.Printf("validateCommitCertificate ::digest1 equality cnt %d", cnt)
	if cnt < F+1 {
		return false
	}
	return true
}
