package server

import (
	"context"
	"fmt"
	"log"
	"time"
	"tpc/db"
	pb "tpc/proto"
)

func (s *TPCServer) TPCPrepare(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	// log.Printf("TPCPrepare inpit length %d : %v", len(in.TransactionEntry), in)
	for index, txn := range in.TransactionEntry {
		res, _ := s.ProcessTransaction(index, txn)
		if !res.Success {
			resp.Success = false
			resp.Message = "Failed in Paxos ::" + res.Message
			return resp, fmt.Errorf(resp.Message)
		}
	}
	return resp, nil

}

func (s *TPCServer) TPCCommit(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	for index, tx := range in.TransactionEntry {
		//res, err := s.ProcessTransaction(index, txn)
		defer s.UnlockResources(tx, index)
		err := s.WALStore.UpdateEntry(index, db.COMMIT, s.ServerID)
		if err != nil {
			resp.Success = false
			resp.Message = "Failed in UpdateEntry" + err.Error()
			return resp, nil
		}
	}
	return resp, nil

}

func (s *TPCServer) TPCAbort(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	resp := &pb.TransactionResponse{
		Success: true,
		Message: "Success",
	}
	for index, tx := range in.TransactionEntry {
		//res, err := s.ProcessTransaction(index, txn)
		log.Printf("On Server %s TPCAbort for txn %d", s.ServerID, index)
		defer s.UnlockResources(tx, index)
		s.RollbackLogs(index, tx)
		// log.Printf("On Server %s TPCAbort for txn %d Rollback done", s.ServerID, index)
		err := s.WALStore.UpdateEntry(index, db.ABORT, s.ServerID)
		log.Printf("On Server %s TPCAbort for txn %d Rollback done , UpdateEntry done", s.ServerID, index)
		if err != nil {
			if err.Error() == db.TXN_NOT_FOUND {
				s.AddToWAL(index, tx, db.ABORT)
				resp.Message = fmt.Sprintf("Success for txn %d", index)
			} else {
				msg := "Failed in UpdateEntry" + err.Error()
				return nil, fmt.Errorf(msg)
			}
		}
	}
	return resp, nil
}

func (s *TPCServer) AddToWAL(index int64, tx *pb.Transaction, status string) {
	log.Printf("AddToWAL on %s for txn %d = %s", s.ServerID, index, status)
	if ShardMap[tx.Sender] != ShardMap[tx.Receiver] {
		entry := &db.WALEntry{
			TransactionID: index,
			Sender:        tx.Sender,
			Receiver:      tx.Receiver,
			Amount:        tx.Amount,
			Status:        status,
			Timestamp:     time.Now().UnixMilli(),
		}
		if err := s.WALStore.AddEntry(entry); err != nil {
			log.Printf("Failed to add transaction %d to WAL: %v", index, err)
		}
	}
}

func (s *TPCServer) ReplayWAL() {
	if err := s.WALStore.ReplayWAL(); err != nil {
		log.Printf("Error replaying WAL: %v", err)
	}
}
