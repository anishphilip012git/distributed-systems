package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tpc/db"
	pb "tpc/proto"
)

func (s *TPCServer) DecidePhaseCall(accpetedTxnMap map[int64]*pb.Transaction) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.peerServers)) // Buffered channel to collect errors

	// Loop through all peers to send Decide requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			// For self, automatically commit logs
			// log.Printf("Committing transactions on server %v", s.ServerID)
			s.CommitLogs(s.acceptedValue)
			s.lastCommittedBallotNumber = max(s.lastCommittedBallotNumber, s.currentBallotNumber)
			// log.Printf("lastCommittedBallotNumber on server %s %d ", s.ServerID, s.lastCommittedBallotNumber)
			log.Println("Decide successful for self")
			continue
		}

		wg.Add(1) // Add to the WaitGroup
		go func(peerId string, peer pb.TPCShardClient) {
			defer wg.Done() // Mark as done when the goroutine finishes
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel() // Ensure resources are freed when the operation completes

			_, err := peer.PaxosDecide(ctx, &pb.DecisionRequest{
				TransactionsEntries: accpetedTxnMap,
				BallotNumber:        s.currentBallotNumber,
			})
			if err != nil {
				// Send error to channel if decide call fails
				errCh <- fmt.Errorf("Decide phase failed on peer %s: %w", peerId, err)
			}
		}(peerId, peer)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh) // Close the error channel after all goroutines finish

	// Aggregate errors
	var aggregatedError error
	for err := range errCh {
		if aggregatedError == nil {
			aggregatedError = err
		} else {
			aggregatedError = fmt.Errorf("%w; %v", aggregatedError, err)
		}
	}

	return aggregatedError
}

// Decide phase: commit the transaction block and clean up logs
func (s *TPCServer) PaxosDecide(ctx context.Context, req *pb.DecisionRequest) (*pb.DecisionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// log.Printf("Committing transactions on server %s %v ", s.ServerID, req.TransactionsEntries)
	// Data store logs

	s.CommitLogs(req.TransactionsEntries)
	for k, _ := range req.TransactionsEntries {
		_, ok := s.uncommittedLogs[k]
		if ok {
			delete(s.uncommittedLogs, k)
		}
	}
	s.lastCommittedBallotNumber = max(s.lastCommittedBallotNumber, req.BallotNumber)
	// log.Printf("lastCommittedBallotNumber on server %s %d ", s.ServerID, s.lastCommittedBallotNumber)
	response := &pb.DecisionResponse{
		Success: true,
	}
	return response, nil
}

func (s *TPCServer) CommitLogs(logs map[int64]*pb.Transaction) {

	if s.committedLogs == nil {
		s.committedLogs = map[int64]*pb.Transaction{}

	}

	for k, txn := range logs {
		isaborted, _ := s.WALStore.CheckEntryStatus(k, db.ABORT)
		// if err != nil {
		// 	log.Printf("CheckEntryStatus on %s for %d ", s.ServerID, k)
		// }
		if isaborted {
			log.Printf("CheckEntryStatus already aborted on %s for %d ", s.ServerID, k)
			continue
		}
		// log.Printf("GetRecord [%d] on server [%s]: \n", k, s.serverID)
		senderBal, senderExists := s.BalanceMap.Read(txn.Sender)
		recvrBal, rcvrExists := s.BalanceMap.Read(txn.Receiver)
		val, _ := db.GetRecord(s.db, db.TBL_EXEC, k)
		if val != "" {
			//log.Printf("Already executed log %d on server %s: %+v , balance on Sender :%d on rcvr:%d\n", k, s.ServerID, txn, senderBal, recvrBal)
			continue
		}

		// log.Printf("Before :: log %d on server %s: %+v , balance on Sender :%d on rcvr:%d \n", k, s.ServerID, txn, senderBal, recvrBal)
		TPCServerMapLock.Lock()
		// s.BalanceMap[txn.Receiver] += int(txn.Amount)
		if senderExists {
			senderBal, _ = s.BalanceMap.Set(txn.Sender, senderBal-int(txn.Amount), k)
		}
		if rcvrExists {
			recvrBal, _ = s.BalanceMap.Set(txn.Receiver, recvrBal+int(txn.Amount), k)
		}

		// s.BalanceMap[txn.Sender] -= int(txn.Amount)
		TPCServerMapLock.Unlock()

		s.committedLogs[k] = txn
		delete(s.uncommittedLogs, k)

		log.Printf("Executing log %d on server %s: %+v ,  balance on Sender ->%d on rcvr:->%d \n", k, s.ServerID, txn, senderBal, recvrBal)
		err := db.StoreRecord(s.db, db.TBL_EXEC, k, txn)
		if err != nil {
			log.Println("Error in committing ", err)
		}

		s.AddToWAL(k, txn, db.PREPARE)
	}

}
func (s *TPCServer) RollbackLogs(index int64, txn *pb.Transaction) {
	_, ok := s.committedLogs[index]
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
		defer delete(s.committedLogs, index)
		defer delete(s.uncommittedLogs, index)
		defer delete(s.acceptedValue, index)
	} else {
		log.Printf("RollbackLogs:: for txn %d : on %s: NOT FOUND", index, s.ServerID)
	}
	err := db.DeleteRecord(s.db, db.TBL_EXEC, index)
	if err != nil {
		log.Printf("RollbackLogs::Error in Deleting on %s :%v", s.ServerID, err)
	}

}
