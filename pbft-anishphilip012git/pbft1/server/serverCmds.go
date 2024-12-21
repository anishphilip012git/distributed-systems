package server

import (
	"fmt"
	"log"
	"pbft1/db"
	"pbft1/metric"
)

// func (s *server.PBFTSvr) PrintBalance(client string) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	//balance, ok := s.balance[client]
// 	balance := s.balance
// 	// /if ok {
// 	log.Printf("Balance of %s on server %s: %d\n", client, s.serverID, balance)
// 	// } else {
// 	// 	log.Printf("Client %s not found on server %s\n", client, s.serverID)
// 	// }
// }

func (s *PBFTSvr) PrinStatus(seqNo int64) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	//balance, ok := s.balance[client]
	output := fmt.Sprintf("on Server %s", s.ServerID)
	if seqNo == -1 {
		for idx, tx := range s.TxnStateBySeq {
			output += fmt.Sprintf(" [%d: %s]", idx, tx)
		}
	} else {
		tx := s.TxnStateBySeq[seqNo]
		output += fmt.Sprintf(" [%v: %s]", (seqNo), tx)
	}
	log.Println(output)
	// } else {
	// 	log.Printf("Client %s not found on server %s\n", client, s.serverID)
	// }
}

func (s *PBFTSvr) PrintLog() {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	log.Printf("Log for server %s:\n", s.ServerID)
	for _, tx := range s.Buffer {
		log.Printf("B: INdex: %d Transaction: %s -> %s, Amount: %d\n", tx.Index, tx.Sender, tx.Receiver, tx.Amount)
	}
	for index, tx := range s.PrePrepareLogs {
		log.Printf("PP: =>: %d -> SequenceNumber %d, ViewNumber: %d : Request: %v \n", index, tx.SequenceNumber, tx.ViewNumber, tx.Request)
	}
	for index, tx := range s.PrepareLogs {
		log.Printf("PP: SequenceNumber : %d Transaction:  %v \n", index, tx)
	}
	for index, tx := range s.CommitedLogs {
		log.Printf("PP: SequenceNumber: %d Transaction:  %v \n", index, tx)
	}

}
func PrintDB(serverID string) {

	log.Printf("Datastore for server %s:\n", serverID)
	// for client, balance := range s.balance {
	// 	log.Printf("%s: %d\n", client, balance)
	// }
	record, err := db.ScanTable(db.DbConnections[serverID], db.TBL_EXEC)
	if err != nil {
		log.Println("Error in getting datastore")
	}
	output := ""
	for index, txn := range record {
		output += fmt.Sprintf("[%d: %s] , ", index, txn)
	}
	log.Println("PrintDB output:", output)
}

func (s *PBFTSvr) Performance() {
	// Placeholder for performance metrics
	metric.PrintMetricsForAllServersAndHandlers(s.ServerID)
}

func (s *PBFTSvr) PrintView() {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	log.Printf("Log for server %s:\n", s.ServerID)
	for index, tx := range s.viewChangeResponses {
		log.Printf("PP: SequenceNumber : %d Transaction:  %v \n", index, tx)
	}

}
