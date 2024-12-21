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

func (s *TPCServer) ValidateCommitMsg(primary string, commmitMsgdata *pb.CommitRequest) (bool, error) {
	sign := commmitMsgdata.LeaderSignature
	if commmitMsgdata.ViewNumber < 0 || commmitMsgdata.SequenceNumber <= 0 {
		return false, fmt.Errorf("Invalid Sequence No")
	}
	commmitMsgdata.LeaderSignature = ""
	// log.Printf("ValidateCommitMsg to verify :: %v ", commmitMsgdata)
	isValid, err := VerifyData(commmitMsgdata, sign, s.PubKeys[primary])
	if !isValid || err != nil {
		msg := fmt.Sprintf("PrePrepare rejected due to incorrect message %v", err)
		return false, fmt.Errorf(msg)
	}
	isValid, err = s.ValidateCommitCertificate(commmitMsgdata)
	// prepMsgData.LeaderSignature = sign
	return isValid, err
}

func (s *TPCServer) ValidateCommitCertificate(commitMsgdata *pb.CommitRequest) (bool, error) {
	msgTotest := commitMsgdata.Digest
	validCnt := 0
	for nodeid, signStr := range commitMsgdata.Signatures {
		isValid, _ := VerifyData(msgTotest, signStr, s.PubKeys[nodeid])
		if isValid {
			validCnt++
		} else {
			log.Printf("ValidatePreparCertificate::Invalid signature in prepare certificate for %s", nodeid)
		}
	}

	if validCnt >= 2*F+1 {
		return true, nil
	} else {
		msg := fmt.Sprintf("ValidateCommitCertificate::Not enough Valid Certificates: %d", validCnt)
		log.Printf(msg)
		return false, fmt.Errorf(msg)
	}

}

// CommitPhaseCall: Executes the request and sends a reply to the client
func (s *TPCServer) CommitPhaseCall(clientRequest *pb.ClientRequestMessage, commitRequest *pb.CommitRequest, seqNo int64) (*pb.StatusResponse, error) {
	log.Println("CommitPhaseCall ::")

	commitCnt := 0

	var wg sync.WaitGroup // To wait for all goroutines to finish
	mutex := sync.Mutex{} // Mutex to safely update execCnt

	// Loop through all peers to send Commit requests
	for peerId, peer := range s.peerServers {
		if ServerClusters[PeerIdMap[peerId]] != ServerClusters[s.ServerID] {
			continue
		}
		if peerId == PeerAddresses[s.ServerID] {
			// s.Mu.Lock()
			// For self, automatically count the commit and log it
			s.TxnState[clientRequest.Transaction.Index] = C_STATE
			s.TxnStateBySeq[seqNo] = C_STATE
			commitCnt++ // Count commit for self
			s.CommittedTxns[clientRequest.Transaction.Index] = clientRequest.Transaction
			clientRequest.Transaction.SeqNo = seqNo
			executed := s.ExecuteTxn(clientRequest.Transaction)
			log.Println("Commit successful for self", executed)
			// s.Mu.Unlock()

			continue
		}

		// Spawning a goroutine for each peer to send the commit request
		if LiveSvrMap[PeerIdMap[peerId]] && !ByzSvrMap[PeerIdMap[peerId]] {
			wg.Add(1)
			go func(peer pb.TPCShardClient, peerId string) {
				defer wg.Done() // Ensure the WaitGroup counter is decremented when done

				// Create context with deadline
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
				defer cancel()

				// Send Commit request to the peer
				resp, err := peer.PBFTCommit(ctx, commitRequest)
				if err != nil {
					log.Printf("Commit failed for %s : %v", s.ServerID, err)
					return // Skip if the peer rejects the Commit request
				}

				// If the response is successful, increment the commit count
				if resp.Success {
					mutex.Lock() // Lock to safely update execCnt
					commitCnt++
					mutex.Unlock()
					log.Printf("Commit successful for %s :: response- %v", s.ServerID, resp)
				}
			}(peer, peerId) // Pass peer and peerId to the goroutine
		}
	}

	// Wait for all goroutines to finish before proceeding
	wg.Wait()

	// After all commits are processed, check if the execution was successful
	statusResponse := &pb.StatusResponse{Success: commitCnt >= (2*F + 1)} // Quorum reached if > len/2

	if commitCnt < (2*F + 1) {
		return statusResponse, fmt.Errorf("Commit failed: Quorum not reached")
	}

	// Send ReplyMessage back to client (this would typically be a callback or message queue)
	log.Printf("Commit successful for Transaction ID %d at server %s", clientRequest.Transaction.Index, s.ServerID)
	return statusResponse, nil
}

func (s *TPCServer) ChkTxnIdExecuted(idx int64) bool {
	chk1, ok := s.TxnState[idx]
	if !ok || chk1 != E_STATE {
		return false
	}
	if ok && chk1 == E_STATE {
		return true
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
func (s *TPCServer) ExecuteTxn(txn *pb.Transaction) bool {

	// idx := txn.Index
	// start := time.Now()
	// Check if transaction is already executed
	execute := false
	if ShardMap[txn.Sender] == ShardMap[txn.Receiver] {
		execute = true
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	for seqno, v := range s.PrePrepareLogs {
		txn := v.Request.Transaction
		txn.SeqNo = seqno
		if s.ChkTxnIdExecuted(txn.Index) {
			log.Printf("ExecuteTxn on %  :: Txn %d: %v already executed", s.ServerID, txn.Index, v)
			continue
		}
		isCommitted, _ := s.WALStore.CheckEntryStatus(txn.Index, db.COMMIT)
		if execute || isCommitted {
			s.markTransactionExecuted(txn)
			delete(s.CommittedTxns, txn.Index)
			delete(s.PrePrepareLogs, seqno)
			log.Printf("ExecuteTxn :: at %s Txn %d: %v  executed", s.ServerID, txn.Index, v)
		}
	}

	for k, v := range s.CommittedTxns {
		if s.ChkTxnIdExecuted(k) {
			log.Printf("ExecuteTxn on %  :: Txn %d: %v already executed", s.ServerID, k, v)
			continue
		}
		isCommitted, _ := s.WALStore.CheckEntryStatus(k, db.COMMIT)
		if execute || isCommitted {
			s.markTransactionExecuted(v)
			delete(s.CommittedTxns, k)
			log.Printf("ExecuteTxn :: at %s Txn %d: %v  executed", s.ServerID, k, v)
		}
	}

	return false
}

// markTransactionExecuted updates the execution state of a transaction and logs it
func (s *TPCServer) markTransactionExecuted(txn *pb.Transaction) {
	idx := txn.Index
	senderBal, senderExists := s.BalanceMap.Read(txn.Sender)
	recvrBal, rcvrExists := s.BalanceMap.Read(txn.Receiver)
	if senderExists {
		senderBal, _ = s.BalanceMap.Set(txn.Sender, senderBal-int(txn.Amount), idx)
	}
	if rcvrExists {
		recvrBal, _ = s.BalanceMap.Set(txn.Receiver, recvrBal+int(txn.Amount), idx)
	}

	s.TxnState[idx] = E_STATE
	s.TxnStateBySeq[txn.SeqNo] = E_STATE
	db.StoreRecord(s.db, db.TBL_EXEC, idx, txn)
	log.Printf("Txn %d: %v executed at server %s", idx, txn, s.ServerID)
}

// // isCheckpointNeeded checks if a checkpoint should be created
// func (s *TPCServer) isCheckpointNeeded(seqNo int64) bool {
// 	return seqNo%s.ChkPointSize == 0
// }
// AppendToPrepareLogs records prepare messages for a specific sequence number

// ExecuteRequest handles the execution of the client request after consensus is reached
func (s *TPCServer) PBFTCommit(ctx context.Context, msg *pb.CommitRequest) (*pb.StatusResponse, error) {
	// 1. Check if the CommitMessage is valid
	log.Printf("PBFTPrepare :: Received prepare message for seqNo %d", msg.SequenceNumber)
	primary := ClusterMapping[ServerClusters[s.ServerID]].Leader
	isValid, err := s.ValidateCommitMsg(primary, msg)
	if err != nil {
		log.Printf("ValidatePrePrepareMsg  Error %v", err)
	}
	if !isValid {
		return nil, fmt.Errorf("Invalid Prepare message")
	}

	var idx int64
	clientreq, exists := DigestMap[msg.Digest]
	if exists {
		idx = clientreq.Transaction.Index
	} else {
		return nil, fmt.Errorf("Invalid Prepare message")
	}

	// 3. Apply the transaction to the replica's state
	// You might use the `in.CommitCertificate` or `in.CommitMessage` to fetch the actual transaction details.
	transaction := DigestMap[msg.Digest].Transaction // Assuming we're using the first PrepareMessage's transaction
	seqNo := msg.SequenceNumber
	transaction.SeqNo = seqNo
	// s.Mu.Lock()
	s.TxnState[idx] = C_STATE
	s.TxnStateBySeq[seqNo] = C_STATE
	s.CommittedTxns[transaction.Index] = transaction
	executed := s.ExecuteTxn(transaction)
	// s.resetViewChangeTimer()
	// s.Mu.Unlock()

	// if err != nil {
	// 	log.Printf("Failed to apply transaction: %v", err)
	// 	return &pb.StatusResponse{Success: false}, err
	// }
	// Reply phase: Replica sends reply to client
	if s.ServerID != ClusterMapping[ServerClusters[s.ServerID]].Leader { //s.id != s.getPrimaryIdForView(msg.ViewNumber)
		log.Printf("PBFTCommit :: CreateClientReplyMessage")
		reply := s.CreateClientReplyMessage(s.TxnState[idx], msg.ViewNumber, idx)
		if ShardMap[transaction.Sender] == ShardMap[transaction.Receiver] {
			go s.SendReplyToClient(reply)
		}

	}
	// 4. Return success
	log.Printf("CommitRequest successful transaction with sequence number %d on Server %s", msg.SequenceNumber, s.ServerID)
	return &pb.StatusResponse{Success: executed}, nil
}
func (s *TPCServer) SendReplyToClient(reply *pb.ClientReplyMessage) {
	start := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeoutForClient))
	defer cancel() // Ensure resources are freed when the operation completes

	// log.Printf("SendReplyToClient for %s %v %v", s.ServerID, s.clientServer, s.clientConnection)
	resp, err := s.clientServer.ReplyMessage(ctx, reply)
	if err != nil {
		log.Printf("SendReplyToClient failed for %s Time taken in failure : %v, err %v", s.ServerID, time.Now().Sub(start).Seconds(), err)
		return
	}
	// log.Printf("SendReplyToClient ::  Time taken at %s : %v", s.ServerID, time.Now().Sub(start).Seconds())
	log.Printf("SendReplyToClient :: passed for server %s with %v , time: %v", s.ServerID, resp, time.Now().Sub(start).Seconds())
}
func (s *TPCServer) CreateClientReplyMessage(msg string, viewNumber, idx int64) *pb.ClientReplyMessage {
	// digest, _ := HashStruct(msg.Request)
	reply := &pb.ClientReplyMessage{
		ViewNumber:       viewNumber,
		Status:           msg,
		RequestId:        idx,
		NodeId:           s.ServerID,
		ReplicaSignature: "",
	}
	sign, _ := SignData(reply, s.PrivateKey)
	reply.ReplicaSignature = sign
	return reply
}
