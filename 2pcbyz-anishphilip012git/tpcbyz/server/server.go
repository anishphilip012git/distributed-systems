package server

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"sync"
	"tpcbyz/db"
	pb "tpcbyz/proto" // Import the generated proto code
	. "tpcbyz/security"
	. "tpcbyz/utils"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

var (
	F                       int
	ClientAddressMapSvrSide map[string]string
	// TxnIdMapWithSeqNo       map[int64]int64
	TxnIdMap           map[int64]*pb.Transaction
	LastCommittedSeqNo = make(map[int]int64)
)

type TPCServer struct {
	pb.UnimplementedTPCShardServer
	Mu             sync.Mutex
	id             int64
	ServerID       string
	addr           string
	ViewNumber     int64
	SequenceNumber int64
	ReplicaCount   int64
	f              int64
	// acceptedBallotNumber      int64
	// lastCommittedBallotNumber int64
	// acceptedValue             map[int64]*pb.Transaction // Changed to handle transactions
	// uncommittedLogs           map[int64]*pb.Transaction // Uncommitted logs now hold a list of transactions
	// clusterSize               int
	peerServers     map[string]pb.TPCShardClient // gRPC clients to other servers
	peerConnections map[string]*grpc.ClientConn
	isDown          bool // Server down status
	// committedLogs             map[int64]*pb.Transaction
	// buffer                    map[int64]*pb.Transaction
	// isLeader        bool           // This will indicate if this server is the leader (Proposer)
	BalanceMap *ConcurrentMap //map[int64]int
	db         *bbolt.DB
	WALStore   *db.WALStore

	// TxnStateMu     sync.RWMutex

	PrivateKey *ecdsa.PrivateKey
	PubKeys    map[string][]byte

	clientServer     pb.PBFTClientServiceClient
	clientConnection *grpc.ClientConn
	PrePrepareLogs   map[int64]*pb.PrePrepareMessage

	PrepareLogs map[int64]*pb.PrepareRequest

	// CommitedLogs map[int64]*pb.PrepareReqMessage
	CommittedTxns map[int64]*pb.Transaction
	ExecutedTxns  map[int64]*pb.Transaction
	TxnState      map[int64]string
	TxnStateBySeq map[int64]string
	Buffer        map[int64]*pb.Transaction

	// TSconfig     *TSReplica
	ClientPubKey []byte
}

var StateMap = map[string]string{
	"PP": "Pre-prepared", //(when the server receives a valid pre-prepare message from the	leader)
	"P":  "Prepared",
	"C":  "Committed",
	"E":  "Executed",
	"X":  "No Status",
}
var (
	PP_STATE = "PP"
	P_STATE  = "P"
	C_STATE  = "C"
	E_STATE  = "E"
	X_STATE  = "X"
	B_STATE  = "B"
)

func (s *TPCServer) VerifyClientRequest(req *pb.ClientRequestMessage) bool {
	sign := req.ClientSignature
	req.ClientSignature = ""
	isValid, err := VerifyData(req, sign, ClientKeys.PublicKey)
	// log.Printf("VerifyClientRequest isValid %v", isValid)
	if !isValid || err != nil {
		log.Printf("VerifyClientRequest rejected due to incorrect message %v", err)
		return false
	}
	// req.ClientSignature = sign
	return isValid
}
func (s *TPCServer) getPrimaryIdForView(newView int64) int64 {
	return newView % s.ReplicaCount
}
func (s *TPCServer) ClientRequest(ctx context.Context, req *pb.ClientRequestMessage) (*pb.ClientReplyMessage, error) {
	log.Printf("Txn %d: %v VerifyClientRequest on %s", req.Transaction.Index, req.Transaction, s.ServerID)
	// senderShard, rcvrShard := ShardMap[req.Transaction.Sender], ShardMap[req.Transaction.Receiver]
	if !s.VerifyClientRequest(req) {

		msg := "Failure"
		replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, req.Transaction.Index)
		return replyMessage, nil
	}
	// s.Mu.Lock()
	// defer s.Mu.Unlock()

	// Check if transaction is already executed
	idx := req.Transaction.Index
	TxnIdMap[idx] = req.Transaction
	state, exists := s.TxnState[idx]
	if exists && state == E_STATE {
		log.Printf("Txn %d: %v already executed on %s", req.Transaction.Index, req.Transaction, s.ServerID)
		msg := "Success"
		replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, req.Transaction.Index)
		return replyMessage, nil
	} else if exists {
		log.Printf("Txn  %v already exists at stage %s on %s", req.Transaction, s.TxnState[idx], s.ServerID)
		return nil, fmt.Errorf("REPEAT_CALL:Skipping")
	}
	log.Printf("[%s] vs [%s] ", s.ServerID, ClusterMapping[ServerClusters[s.ServerID]].Leader)

	if s.ServerID == ClusterMapping[ServerClusters[s.ServerID]].Leader { //s.id == primaryIdx {
		log.Printf("Primary %s received client request, starting view change timer for txn %d", s.ServerID, req.Transaction.Index)
		// Start a timer to trigger view change if request is not executed in time
		// s.startViewChangeTimer()
		// Ensure timer stops when request is executed
		// defer s.resetViewChangeTimer()
		_, replyMessage, err := s.InitiatePBFT(ctx, req)
		defer s.UnlockResources(req.Transaction)
		return replyMessage, err
	} else {
		// Forward request to primary if backup
		peerId := ClusterMapping[ServerClusters[s.ServerID]].Leader
		peer := s.peerServers[peerId]
		if LiveSvrMap[PeerIdMap[peerId]] {
			resp, err := peer.ClientRequest(ctx, req)
			return resp, err
		} else {
			return nil, fmt.Errorf("server is not live %s %v", PeerIdMap[peerId], LiveSvrMap[PeerIdMap[peerId]])
		}
	}
}
