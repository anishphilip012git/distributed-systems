package server

import (
	"context"
	"fmt"
	"log"
	"time"
	pb "tpcbyz/proto"
	. "tpcbyz/security"
)

// var MsgWiseNonceQ map[int64]SharedNonce
// var ChkpntMsgNonceQ map[int64]SharedNonce

// Define the timer duration for a view change trigger
const viewChangeTimeout = 10 * time.Second

const rpcCtxDeadlineTimeoutViewChangeCalls = 1 * time.Second
const rpcCtxDeadlineTimeout = 3 * time.Second

const rpcCtxDeadlineTimeoutForClient = 5 * time.Second

// const F = 1

// func (s *TPCServer) IniitiatePBFT(ctx context.Context, clientRequest *pb.ClientRequestMessage) (*pb.ClientReplyMessage, error) {
// 	// s.mu.Lock()
// 	// defer s.mu.Unlock()
// 	prefix := "IniitiatePBFT ::"
// 	// idx := clientRequest.Transaction.Index
// 	msg := ""
// 	s.Mu.Lock()
// 	s.SequenceNumber++
// 	// s.TxnState[idx] = X_STATE
// 	seqNo := s.SequenceNumber
// 	// TxnIdMapWithSeqNo[idx] = seqNo
// 	s.Mu.Unlock()
// 	// Increment commit count for the sequence number
// 	log.Printf(prefix+"for client request  at %s , %v ", s.ServerID, clientRequest)
// 	if !s.isTxnPossible(clientRequest.Transaction){

// 	}
// 	prepareRequest, _, _ := s.PrePreparePhaseCall(clientRequest, seqNo)
// 	if prepareRequest == nil {
// 		msg = "Prepare rejected"
// 		// time.Sleep(viewChangeTimeout)
// 	} else {
// 		clientRequest.Transaction.SeqNo = seqNo
// 		commitRequest, cnt, _ := s.PreparePhaseCall(clientRequest, prepareRequest, seqNo)

// 		if cnt < F+1 {
// 			msg = "Prepare rejected"
// 		}
// 		s.CommitPhaseCall(clientRequest, commitRequest, seqNo)
// 		msg = "Success"
// 	}
// 	if

// 	replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index)
// 	return replyMessage, nil

// }

// EnsureAllSeqNoProcessed checks if all smaller sequence numbers are either processed or skipped
func (s *TPCServer) EnsureAllSeqNoProcessed(seqNo int64, senderShard, rcvrShard int, index int64) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	log.Printf("EnsureAllSeqNoProcessed:: Checking for txn %d seqNo %d at %s", index, seqNo, s.ServerID)

	if seqNo <= 1 {
		return true
	}

	// Iterate over all sequence numbers before `seqNo`
	for i := int64(1); i < seqNo; i++ {
		// Check if sender and receiver shards have committed this seqNo
		if !s.isSeqNoProcessed(i, senderShard, rcvrShard) {
			log.Printf("EnsureAllSeqNoProcessed:: Sequence number %d is not processed for senderShard %d or rcvrShard %d", i, senderShard, rcvrShard)
			return false
		}
	}

	// If no gaps were found, return true
	return true
}

// isSeqNoProcessed checks if a specific sequence number is either committed or skipped
func (s *TPCServer) isSeqNoProcessed(seqNo int64, senderShard, rcvrShard int) bool {
	// Check committed sequence numbers for sender and receiver shards
	if LastCommittedSeqNo[senderShard] >= seqNo || LastCommittedSeqNo[rcvrShard] >= seqNo {
		return true
	}

	// Check if the seqNo is marked as skipped due to failed consensus
	_, ok := s.TxnStateBySeq[seqNo]
	if ok { //SkippedSeqNos[senderShard][seqNo] || SkippedSeqNos[rcvrShard][seqNo] {
		return true
	}

	return false
}

func (s *TPCServer) InitiatePBFT(ctx context.Context, clientRequest *pb.ClientRequestMessage) (*pb.CommitRequest, *pb.ClientReplyMessage, error) {
	prefix := "InitiatePBFT :: at " + s.ServerID + " "
	log.Printf(prefix+"for client request at %s, %v", s.ServerID, clientRequest)

	// Validate transaction feasibility
	senderShard := ShardMap[clientRequest.Transaction.Sender]
	rcvrShard := ShardMap[clientRequest.Transaction.Receiver]
	_, err := s.isTxnPossible(clientRequest.Transaction)
	if err != nil {
		msg := err.Error()
		return nil, s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index), nil
	}
	// Increment sequence number and ensure all smaller sequence numbers are processed
	seqNo := s.incrementAndEnsureSequenceNumbers(senderShard, rcvrShard, clientRequest.Transaction.Index)

	// Intra-shard transaction preprocessing
	// if senderShard == rcvrShard {
	// 	missingSeqNos := s.findMissingSeqNos(s.PrePrepareLogs, s.PrepareLogs, seqNo)
	// 	for _, mseqNo := range missingSeqNos {
	// 		log.Printf(prefix+"missingSeqNos CONSENSUS for SeqNo %d on server %s START", mseqNo, s.ServerID)
	// 		_, ok := s.PrePrepareLogs[int64(mseqNo)]
	// 		if ok {
	// 			oldReq := s.PrePrepareLogs[int64(mseqNo)].Request
	// 			_, msg, success := s.ExecutePBFTConsensus(oldReq, int64(mseqNo))
	// 			log.Printf(prefix+"missingSeqNos CONSENSUS for SeqNo %d on server %s FINISH :result %s success:%s", mseqNo, s.ServerID, msg, success)
	// 			if success {
	// 				s.updateCommittedSeqNo(senderShard, rcvrShard, int64(mseqNo))
	// 			}
	// 		}
	// 	}
	// 	log.Printf(prefix+"at %s missingSeqNos %v", s.ServerID, missingSeqNos)

	// }

	// Execute PBFT consensus phases
	commitRequest, msg, success := s.ExecutePBFTConsensus(clientRequest, seqNo)
	if !success {
		return nil, s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index), nil
	}

	// Handle shard-specific transaction processing
	msg = s.processShardTransaction(clientRequest, senderShard, rcvrShard, seqNo)
	if msg != "Success" {
		return nil, s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index), nil
	}

	// Create reply message
	replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index)
	return commitRequest, replyMessage, nil
}

// Increment the sequence number and ensure smaller sequence numbers are processed
func (s *TPCServer) incrementAndEnsureSequenceNumbers(senderShard, rcvrShard int, txnIndex int64) int64 {
	s.Mu.Lock()
	s.SequenceNumber++
	seqNo := s.SequenceNumber
	s.Mu.Unlock()

	for {
		if s.EnsureAllSeqNoProcessed(seqNo, senderShard, rcvrShard, txnIndex) {
			break
		}
		log.Printf("Waiting for smaller sequence numbers to be processed for transaction %d", seqNo)
		time.Sleep(1000 * time.Millisecond) // Sleep before retrying
	}
	return seqNo
}

// Execute PBFT consensus phases
func (s *TPCServer) ExecutePBFTConsensus(clientRequest *pb.ClientRequestMessage, seqNo int64) (*pb.CommitRequest, string, bool) {
	log.Printf("PBFT STARTED for transaction %d at %s", seqNo, s.ServerID)

	// Pre-Prepare Phase
	prepareRequest, _, _ := s.PrePreparePhaseCall(clientRequest, seqNo)
	if prepareRequest == nil {
		digest, _ := HashStruct(clientRequest)
		commitRequest := &pb.CommitRequest{
			SequenceNumber: seqNo,
			Digest:         digest, ViewNumber: s.ViewNumber,
		}
		msg := "Pre Prepare rejected"
		return commitRequest, msg, false
	}

	// Prepare Phase
	clientRequest.Transaction.SeqNo = seqNo
	commitRequest, cnt, _ := s.PreparePhaseCall(clientRequest, prepareRequest, seqNo)
	if cnt < F+1 {
		msg := "Prepare rejected"
		return commitRequest, msg, false
	}

	// Commit Phase
	s.CommitPhaseCall(clientRequest, commitRequest, seqNo)
	return commitRequest, "Success", true
}
func (s *TPCServer) findMissingSeqNos(prePrepareLogs map[int64]*pb.PrePrepareMessage, prepareLogs map[int64]*pb.PrepareRequest, currSeqNo int64) []int64 {
	var missingSeqNos []int64

	// Find the highest seqNo in PrepareLogs
	maxPrepareSeqNo := int64(0)
	for seqNo := range prepareLogs {
		if seqNo > maxPrepareSeqNo {
			maxPrepareSeqNo = seqNo
		}
	}
	log.Println("findMissingSeqNos :: maxPrepareSeqNo", maxPrepareSeqNo)
	// Iterate through PrePrepareLogs to check for missing entries in PrepareLogs
	log.Printf("prepareLogs at %s %v ", s.ServerID, prepareLogs)
	log.Printf("prePrepareLogs %v ", prePrepareLogs)
	for seqNo := range prePrepareLogs {
		// Ignore sequence numbers greater than the maxPrepareSeqNo
		if seqNo >= currSeqNo {
			continue
		}

		// Check if seqNo is missing in PrepareLogs
		if _, exists := prepareLogs[seqNo]; !exists {
			missingSeqNos = append(missingSeqNos, seqNo)
		}
	}

	return missingSeqNos
}

// Process shard-specific transaction logic
func (s *TPCServer) processShardTransaction(clientRequest *pb.ClientRequestMessage, senderShard, rcvrShard int, seqNo int64) string {
	prefix := "InitiatePBFT :: "
	if senderShard == rcvrShard {

		// Intra-shard transaction
		log.Printf(prefix+"Intra-shard transaction detected at %s %v", s.ServerID, clientRequest)
		s.updateCommittedSeqNo(senderShard, rcvrShard, seqNo)
	} else {
		// Cross-shard transaction
		log.Printf(prefix+"Cross-shard transaction detected at %s %v", s.ServerID, clientRequest)
		if ServerClusters[s.ServerID] == senderShard {
			if !s.ProcessCrossShardTxn(senderShard, rcvrShard, clientRequest) {
				return "Cross-shard transaction processing failed"
			}
			s.updateCommittedSeqNo(senderShard, rcvrShard, seqNo)
		}
	}
	return "Success"
}

// Update the last committed sequence numbers for the shards
func (s *TPCServer) updateCommittedSeqNo(senderShard, rcvrShard int, seqNo int64) {
	if LastCommittedSeqNo[senderShard] < seqNo {
		LastCommittedSeqNo[senderShard] = seqNo
	}
	if LastCommittedSeqNo[rcvrShard] < seqNo {
		LastCommittedSeqNo[rcvrShard] = seqNo
	}
}

// ProcessCrossShardTxn handles transactions across shards
func (s *TPCServer) ProcessCrossShardTxn(senderShard, rcvrShard int, in *pb.ClientRequestMessage) bool {
	log.Printf("ProcessCrossShardTxn :: Processing cross-shard transaction from shard %d to shard %d", senderShard, rcvrShard)

	// Add TPC logic for cross-shard transaction
	if !s.TPCPrepareCall(in) {
		log.Println("ProcessCrossShardTxn :: Cross-shard TPC prepare phase failed")
		return false
	}
	return s.TPCCommitCall(in)
}
func (s *TPCServer) isTxnPossible(tx *pb.Transaction) (bool, error) {
	index := tx.Index

	// Maximum retries for acquiring the lock
	maxRetries := 5
	// Delay between retries
	retryDelay := 500 * time.Millisecond

	// Check if sender is part of the current server's shard
	if ServerClusters[s.ServerID] == ShardMap[tx.Sender] {
		// Try acquiring lock on sender's account
		var senderBalance int
		var err error

		for retries := 0; retries < maxRetries; retries++ {
			senderBalance, err = s.BalanceMap.GetAndLock(tx.Sender, index, s.ServerID)
			if err == nil {
				break // Lock acquired successfully
			}
			log.Printf("Attempt %d to lock sender %d failed: %v", retries+1, tx.Sender, err)
			time.Sleep(retryDelay) // Wait before retrying
		}

		// If lock could not be acquired after retries
		if err != nil {
			errMsg := fmt.Sprintf(err.Error()+": tx %d : Sender %d, Amount %d", index, tx.Sender, tx.Amount)
			log.Println(errMsg)
			return false, fmt.Errorf("LOCK_UNAVAILABLE")
		}

		log.Printf("BALANCE CHK:: balance %d <> amount %d ", senderBalance, int(tx.Amount))
		if senderBalance < int(tx.Amount) {

			return false, fmt.Errorf("INSUFFICIENT_BAL")

		}
	}

	// Check if receiver is part of the current server's shard
	if ServerClusters[s.ServerID] == ShardMap[tx.Receiver] {
		// Try acquiring lock on receiver's account
		for retries := 0; retries < maxRetries; retries++ {
			_, err := s.BalanceMap.GetAndLock(tx.Receiver, index, s.ServerID)
			if err == nil {
				break // Lock acquired successfully
			}
			log.Printf("Attempt %d to lock receiver %d failed: %v", retries+1, tx.Receiver, err)
			time.Sleep(retryDelay) // Wait before retrying
		}
	}

	return true, nil
}
