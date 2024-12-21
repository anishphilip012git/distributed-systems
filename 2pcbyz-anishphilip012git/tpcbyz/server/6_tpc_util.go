package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tpcbyz/db"
	pb "tpcbyz/proto"
	. "tpcbyz/security"
)

var TxnCommitRequest = make(map[int64]*pb.CommitRequest)

// Perform PBFT Step
func (s *TPCServer) performPBFT(clientRequest *pb.ClientRequestMessage) (*pb.CommitRequest, error) {
	// Validate transaction feasibility
	senderShard := ShardMap[clientRequest.Transaction.Sender]
	rcvrShard := ShardMap[clientRequest.Transaction.Receiver]
	_, err := s.isTxnPossible(clientRequest.Transaction)
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf(msg)
	}
	// Increment sequence number and ensure all smaller sequence numbers are processed
	seqNo := s.incrementAndEnsureSequenceNumbers(senderShard, rcvrShard, clientRequest.Transaction.Index)

	// Execute PBFT consensus phases
	commitRequest, msg, success := s.ExecutePBFTConsensus(clientRequest, seqNo)
	if !success {
		return commitRequest, fmt.Errorf(msg)
	}
	return commitRequest, nil
}

// Broadcast Commit Request to Relevant Peers
// func (s *TPCServer) broadcastCommit(ctx context.Context, in *pb.ClientRequestMessage, commitRequest *pb.CommitRequest, messageType string) {
// 	var wg sync.WaitGroup
// 	tpcCerti := &pb.TPCCertificate{
// 		CommitRequest: commitRequest,
// 		Message:       messageType,
// 	}
// 	senderShard := ShardMap[in.Transaction.Sender]

// 	for peerId, peer := range s.peerServers {
// 		log.Printf("broadcastCommit %s %d %d", peerId, ServerClusters[PeerIdMap[peerId]], (senderShard))
// 		// Skip peers not in the sender shard's cluster
// 		if ServerClusters[PeerIdMap[peerId]] != (senderShard) {
// 			continue
// 		}

// 		// Skip peers that are not live
// 		if !LiveSvrMap[PeerIdMap[peerId]] {
// 			continue
// 		}
// 		wg.Add(1)
// 		// go
// 		func(peer pb.TPCShardClient, peerId string) {
// 			defer wg.Done()
// 			// Create context with deadline
// 			// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
// 			// defer cancel()
// 			log.Printf("broadcastCommit :: at peer %s ", PeerIdMap[peerId])
// 			// Send the TPC certificate to the peer
// 			resp, err := peer.TPCCooridnatorCall(ctx, tpcCerti)
// 			if err != nil {
// 				log.Printf("broadcastCommit :: TPCPrepared failed for peer %s: %v", peerId, err)
// 				return
// 			}
// 			if resp.Success {
// 				log.Printf("broadcastCommit :: TPCPrepared successful for peer %s with message type %s", PeerIdMap[peerId], messageType)
// 			}
// 		}(peer, peerId) // Pass peer and peerId to the goroutine
// 	}

// 	wg.Wait() // Wait for all broadcasts to complete
// 	log.Printf("Broadcast completed for message type %s at %s", messageType, s.ServerID)

// }

// Broadcast Commit Request to Relevant Peers
func (s *TPCServer) broadcastCordPart(in *pb.ClientRequestMessage, commitRequest *pb.CommitRequest, messageType string) {
	var wg sync.WaitGroup
	if commitRequest == nil {
		digest, _ := HashStruct(in)
		commitRequest = &pb.CommitRequest{
			SequenceNumber: in.Transaction.SeqNo,
			Digest:         digest,
			ViewNumber:     s.ViewNumber,
		}
	}
	tpcCerti := &pb.TPCCertificate{
		CommitRequest: commitRequest,
		Message:       messageType,
	}
	senderShard := ShardMap[in.Transaction.Sender]
	rcvrShard := ShardMap[in.Transaction.Receiver]

	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			log.Printf("broadcastCordPart %v , %v, %v ", in.Transaction, messageType, commitRequest.ViewNumber)
			s.HandleMessage(in.Transaction, messageType, commitRequest.ViewNumber)
			continue
		}
		// Skip peers not in the sender shard's cluster
		if !(ServerClusters[PeerIdMap[peerId]] == (senderShard) || ServerClusters[PeerIdMap[peerId]] == (rcvrShard)) {
			continue
		}

		// Skip peers that are not live
		if !LiveSvrMap[PeerIdMap[peerId]] {
			continue
		}

		wg.Add(1)
		// go
		func(peer pb.TPCShardClient, peerId string) {
			defer wg.Done()
			// Create context with deadline
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel()
			// Send the TPC certificate to the peer
			resp, err := peer.TPCCooridnatorCall(ctx, tpcCerti)
			if err != nil {
				log.Printf("broadcastCordPart :: TPCCooridnatorCall failed for peer %s: %v", PeerIdMap[peerId], err)
				return
			}
			if resp.Success {
				log.Printf("broadcastCordPart :: TPCCooridnatorCall successful for peer %s with message type %s", PeerIdMap[peerId], messageType)
			}
		}(peer, peerId) // Pass peer and peerId to the goroutine
	}

	wg.Wait() // Wait for all broadcasts to complete
	log.Printf("Broadcast completed for message type %s at %s", messageType, s.ServerID)

}

func (s *TPCServer) UpdateWAL(txn *pb.Transaction, walMsg string) (string, error) {
	msg := ""
	err := s.WALStore.UpdateEntry(txn.Index, walMsg, s.ServerID)
	log.Printf("On Server %s TPCAbort for txn %d UpdateEntry done", s.ServerID, txn.Index)
	if err != nil {
		if err.Error() == db.TXN_NOT_FOUND {
			s.AddToWAL(txn, walMsg)
			msg = fmt.Sprintf("Success for txn %d", txn.Index)
		} else {
			msg := "Failed in UpdateEntry" + err.Error()
			return "", fmt.Errorf(msg)
		}
	}
	return msg, nil
}
func (s *TPCServer) TPCCooridnatorCall(ctx context.Context, in *pb.TPCCertificate) (*pb.TransactionResponse, error) {
	viewNo := int64(0)
	if in.CommitRequest != nil {
		viewNo = in.CommitRequest.ViewNumber
	}

	req, _ := DigestMap[in.CommitRequest.Digest]
	// if !ok {
	// 	log.Println("in.CommitRequest.Digest", in.CommitRequest)
	// }
	txn := req.Transaction
	log.Printf("TPCCooridnatorCall :: on %s for txn %v ", s.ServerID, txn)
	resp, err := s.HandleMessage(txn, in.Message, viewNo)
	return resp, err

}
func (s *TPCServer) HandleMessage(txn *pb.Transaction, msg string, viewNo int64) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	var err error
	switch msg {
	case db.PREPARE:
		resp.Message, err = s.UpdateWAL(txn, msg)
	case db.ABORT:
		resp.Message, err = s.UpdateWAL(txn, msg)
		s.RollbackLogs(txn)
		defer s.UnlockResources(txn)
	case db.COMMIT:
		s.WALStore.UpdateEntry(txn.Index, db.COMMIT, s.ServerID)
		s.ExecuteTxn(txn)
		log.Printf("TPCCooridnatorCall :: HandleMessage :: ExecuteTxn")
		reply := s.CreateClientReplyMessage(s.TxnState[txn.Index], viewNo, txn.Index)
		s.SendReplyToClient(reply)
		defer s.UnlockResources(txn)
	default:
		log.Printf("TPCCooridnatorCall :: Invalid Message for txn %v ", txn)
	}
	return resp, err

}

func (s *TPCServer) AddToWAL(tx *pb.Transaction, status string) {
	log.Printf("AddToWAL on %s for txn %d = %s", s.ServerID, tx.Index, status)
	if ShardMap[tx.Sender] != ShardMap[tx.Receiver] {
		entry := &db.WALEntry{
			TransactionID: tx.Index,
			Sender:        tx.Sender,
			Receiver:      tx.Receiver,
			Amount:        tx.Amount,
			Status:        status,
			Timestamp:     time.Now().UnixMilli(),
		}
		if err := s.WALStore.AddEntry(entry); err != nil {
			log.Printf("Failed to add transaction %d to WAL: %v", tx.Index, err)
		}
	}
}

func (s *TPCServer) ReplayWAL() {
	if err := s.WALStore.ReplayWAL(); err != nil {
		log.Printf("Error replaying WAL: %v", err)
	}
}

// UnlockResources releases locks on both sender and receiver after transaction processing
func (s *TPCServer) UnlockResources(tx *pb.Transaction) {
	index := tx.Index
	if ServerClusters[s.ServerID] == ShardMap[tx.Sender] {
		s.BalanceMap.UnlockKey(tx.Sender, index, s.ServerID)
		// log.Printf("UNLOCK Sender's account SUCCESS %d for txn:%d", tx.Sender, index)
	}

	if ServerClusters[s.ServerID] == ShardMap[tx.Receiver] {
		s.BalanceMap.UnlockKey(tx.Receiver, index, s.ServerID)
		// log.Printf("UNLOCK Receiver's account SUCCESS %d for txn:%d", tx.Receiver, index)
	}

}
func (s *TPCServer) RollbackLogs(txn *pb.Transaction) {
	index := txn.Index
	_, ok := s.CommittedTxns[index]
	if ok {
		senderBal, senderExists := s.BalanceMap.Read(txn.Sender)
		recvrBal, rcvrExists := s.BalanceMap.Read(txn.Receiver)
		//TPCServerMapLock.Lock()
		// s.BalanceMap[txn.Receiver] += int(txn.Amount)
		if rcvrExists {
			recvrBal, _ = s.BalanceMap.Set(txn.Receiver, recvrBal-int(txn.Amount), index)
		}
		if senderExists {
			senderBal, _ = s.BalanceMap.Set(txn.Sender, senderBal+int(txn.Amount), index)
		}
		// s.BalanceMap[txn.Sender] -= int(txn.Amount)
		// TPCServerMapLock.Unlock()
		log.Printf("RollbackLogs:: for txn %d : on %s: new senderBal:%d recvrBal:%d ", index, s.ServerID, senderBal, recvrBal)
		defer delete(s.CommittedTxns, index)
		defer delete(s.ExecutedTxns, index)
	} else {
		log.Printf("RollbackLogs:: for txn %d : on %s: NOT FOUND", index, s.ServerID)
	}
	err := db.DeleteRecord(s.db, db.TBL_EXEC, index)
	if err != nil {
		log.Printf("RollbackLogs::Error in Deleting on %s :%v", s.ServerID, err)
	}

}
