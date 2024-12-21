package server

import (
	"context"
	"fmt"
	"log"
	"time"
	"tpcbyz/db"
	pb "tpcbyz/proto"
)

// TPCPrepare simulates the TPC prepare phase
func (s *TPCServer) TPCPrepareCall(in *pb.ClientRequestMessage) bool {
	txn := in.Transaction
	log.Printf("TPC Prepare phase for transaction %d", txn.Index)
	rcvrPrimaryIndex := ClusterMapping[ShardMap[txn.Receiver]].Leader //server.AllServers[c.viewNumber%int64(len(c.replicas))]
	log.Printf("Sending REQUEST to recvr primary %s ", rcvrPrimaryIndex)
	peer := s.peerServers[PeerAddresses[rcvrPrimaryIndex]]
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
	defer cancel()

	// Send Commit request to the peer
	_, err := peer.TPCPrepare(ctx, in)
	if err != nil {
		log.Printf("TPCPrepare failed for %s : %v", s.ServerID, err)
		return false // Skip if the peer rejects the Commit request
	}
	log.Printf("TPCPrepareCall :: returns True")
	return true // Return false if preparation fails
}

func (s *TPCServer) TPCPrepare(ctx context.Context, in *pb.ClientRequestMessage) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	log.Printf("TPCPrepare input at %s %d : %v", s.ServerID, in.Transaction.Index, in)

	// Perform PBFT step
	commitRequest, pbftErr := s.performPBFT(in)
	log.Printf("TPCPrepare PBFT OUTPUT %s %d : %v", s.ServerID, in.Transaction.Index, commitRequest)
	TxnCommitRequest[in.Transaction.Index] = commitRequest
	if pbftErr != nil {
		resp.Success = false
		resp.Message = "TPCPrepare :: Failed in PBFT step TPCPrepare on receiver side: " + pbftErr.Error()
		log.Println(resp.Message)

		// Broadcast Abort to all relevant peers
		s.broadcastCordPart(in, commitRequest, db.ABORT)
		return resp, fmt.Errorf(resp.Message)
	}

	// Broadcast Prepare to all relevant peers
	s.broadcastCordPart(in, commitRequest, db.PREPARE)
	return resp, nil
}

// TPCCommit simulates the TPC commit phase
func (s *TPCServer) TPCCommitCall(in *pb.ClientRequestMessage) bool {
	txn := in.Transaction
	log.Printf("TPCCommitCall :: TPC Commit phase for transaction %d", txn.Index)
	// Add commit logic here

	s.broadcastCordPart(in, TxnCommitRequest[txn.Index], db.COMMIT)

	return true // Return false if commit fails
}

func (s *TPCServer) TPCAbort(ctx context.Context, in *pb.ClientRequestMessage) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	tx, index := in.Transaction, in.Transaction.Index
	//res, err := s.ProcessTransaction(index, txn)
	log.Printf("TPCAbort :: On Server %s TPCAbort for txn %d", s.ServerID, index)
	defer s.UnlockResources(tx)
	s.RollbackLogs(tx)
	log.Printf("TPCAbort :: On Server %s TPCAbort for txn %d Rollback done", s.ServerID, index)
	err := s.WALStore.UpdateEntry(index, db.ABORT, s.ServerID)
	log.Printf("TPCAbort :: On Server %s TPCAbort for txn %d UpdateEntry done", s.ServerID, index)
	if err != nil {
		if err.Error() == db.TXN_NOT_FOUND {
			s.AddToWAL(tx, db.ABORT)
			resp.Message = fmt.Sprintf("Success for txn %d", index)
		} else {
			msg := "Failed in UpdateEntry" + err.Error()
			return nil, fmt.Errorf(msg)
		}
	}
	return resp, nil
}
