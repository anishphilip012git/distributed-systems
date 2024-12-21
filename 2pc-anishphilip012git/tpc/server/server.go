package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tpc/db"
	"tpc/metric"
	pb "tpc/proto" // Import the generated proto code
	. "tpc/utils"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

type TPCServer struct {
	pb.UnimplementedTPCShardServer
	mu                        sync.Mutex
	ServerID                  string
	addr                      string
	currentBallotNumber       int64
	acceptedBallotNumber      int64
	lastCommittedBallotNumber int64
	acceptedValue             map[int64]*pb.Transaction // Changed to handle transactions
	uncommittedLogs           map[int64]*pb.Transaction // Uncommitted logs now hold a list of transactions
	clusterSize               int
	peerServers               map[string]pb.TPCShardClient // gRPC clients to other servers
	peerConnections           map[string]*grpc.ClientConn
	isDown                    bool // Server down status
	committedLogs             map[int64]*pb.Transaction
	buffer                    map[int64]*pb.Transaction
	// isLeader        bool           // This will indicate if this server is the leader (Proposer)
	BalanceMap *ConcurrentMap //map[int64]int
	db         *bbolt.DB
	WALStore   *db.WALStore
}

// Implement the SendTransaction RPC method in TPCServer
func (s *TPCServer) SendTransaction(ctx context.Context, req *pb.TransactionRequest) (returnVal *pb.TransactionResponse, err error) {
	log.SetPrefix("Server " + s.ServerID + "::")

	// Process the transaction here, e.g., validate the transaction, apply to balance, etc.
	for index, tx := range req.TransactionEntry {
		if s.isDown {
			log.Printf("Server %s is down.Buffering transaction", s.ServerID)
			s.buffer[index] = tx
			return &pb.TransactionResponse{
				Success: false,
				Message: "Server is Down",
			}, nil
		} else {
			// if len(s.buffer) > 0 {
			// 	log.Printf("Server %s Buffer is not empty processing buffer first ", s.ServerID)

			// 	// for bufIndex, buffTxn := range s.buffer {
			// 	// 	s.ProcessTransaction(bufIndex, buffTxn)
			// 	// }
			// }
		}
		log.Printf("Server %s received transaction %d from %d to %d of amount %d\n", s.ServerID, index, tx.Sender, tx.Receiver, tx.Amount)
		returnVal, err = s.ProcessTransaction(index, tx)
		// err := s.WALStore.UpdateEntry(index, db.COMMIT)
		defer s.UnlockResources(tx, index)
		if returnVal != nil {
			return returnVal, err
		}
	}
	// Respond with success
	return &pb.TransactionResponse{
		Success: true,
		Message: "Transaction successful",
	}, nil
}

// gRPC Service: Start Paxos and return immediately.
func (s *TPCServer) ProcessTransaction(index int64, tx *pb.Transaction) (*pb.TransactionResponse, error) {
	s.mu.Lock()
	// log.Printf("MUTEX LOCK for tx %d on server %s ", index, s.ServerID)
	// defer log.Printf("MUTEX UNLOCK for tx %d on server %s ", index, s.ServerID)
	defer s.mu.Unlock()
	start := time.Now() // Record the start time
	// if int(s.BalanceMap[tx.Sender]) >= int(tx.Amount) {
	log.Printf("Starting Paxos consensus... on %s for tx %d", s.ServerID, index)
	s.addToUncommittedLogs(index, tx)
	delete(s.buffer, index)
	success, err := s.InitiatePaxosConsensus(index, tx)
	if err != nil || !success {
		log.Printf("Paxos consensus failed: %v", err)
		delete(s.uncommittedLogs, index)
		return &pb.TransactionResponse{
			Success: success,
			Message: fmt.Sprintf("Transaction processed %v. Paxos Failed.", success),
		}, nil
	}

	log.Printf("Paxos: Transaction %d processed successfully on %s", index, s.ServerID)
	// }

	elapsed := time.Since(start) // Calculate the elapsed time
	metric.TnxSum.Mutex.Lock()
	metric.TnxSum.TotalElapsedTime += elapsed
	metric.TnxSum.TransactionCount++
	metric.TnxSum.Mutex.Unlock()
	db.StoreTime(s.db, index, tx, elapsed)
	return &pb.TransactionResponse{
		Success: true,
		Message: fmt.Sprintf("Transaction processed %v. Please check logs for further status.", true),
	}, nil
}

// Helper function to add a transaction to uncommitted logs.
func (s *TPCServer) addToBufferLogs(index int64, tx *pb.Transaction) {
	if s.buffer == nil {
		s.buffer = make(map[int64]*pb.Transaction)
	}
	s.buffer[index] = tx
}

// Helper function to add a transaction to uncommitted logs.
func (s *TPCServer) addToUncommittedLogs(index int64, tx *pb.Transaction) {
	if s.uncommittedLogs == nil {
		s.uncommittedLogs = make(map[int64]*pb.Transaction)
	}
	s.uncommittedLogs[index] = tx
}
