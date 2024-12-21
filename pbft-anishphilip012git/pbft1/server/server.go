package server

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	pb "pbft1/proto" // Import the generated proto code
	. "pbft1/security"
	"sync"
	"time"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

var (
	ReplicaCount            int64
	F                       int
	ClientAddressMapSvrSide map[string]string
	TxnIdMapWithSeqNo       map[int64]int64
	TxnIdMap                map[int64]*pb.Transaction
)

type PBFTSvr struct {
	pb.UnimplementedPBFTServiceServer
	Mu             sync.RWMutex
	TxnStateMu     sync.RWMutex
	ServerID       string
	Addr           string
	id             int64
	ViewNumber     int64
	SequenceNumber int64
	// replicaCount    int64
	// f               int
	BalanceMap      map[string]int
	peerServers     map[string]pb.PBFTServiceClient // gRPC clients to other servers
	peerConnections map[string]*grpc.ClientConn
	PrivateKey      *ecdsa.PrivateKey
	PubKeys         map[string][]byte

	clientServers     map[string]pb.PBFTClientServiceClient
	clientConnections map[string]*grpc.ClientConn
	PrePrepareLogs    map[int64]*pb.PrePrepareMessage

	PrepareLogs map[int64][]*pb.PrepareMessage

	CommitedLogs map[int64]*pb.PrepareReqMessage
	// CommittedTxns   map[int64]*pb.Transaction
	// ExecutedTxns    map[int64]*pb.Transaction
	TxnState      map[int64]string
	TxnStateBySeq map[int64]string
	Buffer        map[int64]*pb.Transaction

	IsDown           bool // Server down status
	TSconfig         *TSReplica
	ClientPubKey     []byte
	db               *bbolt.DB
	stableCheckpoint int64

	// lowWaterMark        int64
	// highWaterMark       int64
	waterMarkRange       int64
	viewChangeTimer      *time.Timer
	ChkPointSize         int64
	Checkpoints          map[int64]*pb.CheckpointMessage
	viewChangeResponses  map[int64][]*pb.ViewChangeMessage
	viewChangeMutex      sync.RWMutex
	viewChangeInProgress int32
	isBusy               bool
	busyLock             sync.RWMutex
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

func (s *PBFTSvr) Reset() {
	s.Mu.Lock()

	s.ViewNumber = 0
	s.SequenceNumber = 0

	s.BalanceMap = DefaultBalanceMap(10)

	s.PrePrepareLogs = map[int64]*pb.PrePrepareMessage{}

	s.PrepareLogs = map[int64][]*pb.PrepareMessage{}

	s.CommitedLogs = map[int64]*pb.PrepareReqMessage{}

	s.TxnState = map[int64]string{}
	s.TxnStateBySeq = map[int64]string{}
	s.Buffer = map[int64]*pb.Transaction{}

	s.stableCheckpoint = 0
	s.Checkpoints = map[int64]*pb.CheckpointMessage{}
	s.viewChangeResponses = map[int64][]*pb.ViewChangeMessage{}
	s.Mu.Unlock()
	MsgWiseNonceQ = map[int64]SharedNonce{}
	ChkpntMsgNonceQ = map[int64]SharedNonce{}
	DigestMap = map[string]*pb.ClientRequestMessage{}
	TxnIdMapWithSeqNo = map[int64]int64{}
	TxnIdMap = map[int64]*pb.Transaction{}
}
func (s *PBFTSvr) AppendToPrepareLogs(prepareMessage *pb.PrepareMessage, seqNo int64) {
	// Check if there is already a slice for the given sequence number
	if s.PrepareLogs == nil {
		s.PrepareLogs = map[int64][]*pb.PrepareMessage{}
	}
	if _, exists := s.PrepareLogs[seqNo]; !exists {
		// Initialize the slice if it doesn't exist
		s.PrepareLogs[seqNo] = []*pb.PrepareMessage{}
	}

	// Append the prepareMessage to the slice at the given sequence number
	s.PrepareLogs[seqNo] = append(s.PrepareLogs[seqNo], prepareMessage)
}
func (s *PBFTSvr) AppendToPrePrepareLogs(prePrepareMsg *pb.PrePrepareMessage, seqNo int64) {
	// Check if there is already a slice for the given sequence number
	if s.PrePrepareLogs == nil {
		s.PrePrepareLogs = map[int64]*pb.PrePrepareMessage{}
	}
	// Append the prepareMessage to the map at the given sequence number
	s.PrePrepareLogs[seqNo] = prePrepareMsg
}

func (s *PBFTSvr) VerifyClientRequest(req *pb.ClientRequestMessage) bool {
	sign := req.ClientSignature
	req.ClientSignature = ""
	isValid, err := VerifyData(req, sign, ClientKeys.PublicKey)
	if !isValid || err != nil {
		log.Printf("VerifyCommitMessage rejected due to incorrect message %v", err)
		return false
	}
	req.ClientSignature = sign
	return isValid
}
func (s *PBFTSvr) ClientRequest(ctx context.Context, req *pb.ClientRequestMessage) (*pb.ClientReplyMessage, error) {
	for {
		if !s.isBusy {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Set the flag to indicate that a request is being processed
	s.busyLock.Lock()
	s.isBusy = true
	s.busyLock.Unlock()
	// Ensure the flag is reset after the request is processed
	defer func() {
		s.busyLock.Lock()
		s.isBusy = false
		s.busyLock.Unlock()
	}()
	log.Printf("Server %s recieved request %v", s.ServerID, req.Transaction)
	if !s.VerifyClientRequest(req) {
		log.Printf("Txn %d: %v already executed on %s", req.Transaction.Index, req.Transaction, s.ServerID)
		// msg := "Failure"
		// replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, req.Transaction.Index)
		return nil, nil
	}
	// Check if transaction is already executed
	idx := req.Transaction.Index
	TxnIdMap[idx] = req.Transaction
	state, exists := s.TxnState[idx]
	if exists && state == E_STATE {
		log.Printf("Txn %d: %v already executed on %s", req.Transaction.Index, req.Transaction, s.ServerID)
		msg := "Success"
		replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, req.Transaction.Index)
		return replyMessage, nil
	}

	// Start the timer only if this is a backup replica (not primary) and the transaction is not yet executed
	primaryIdx := s.getPrimaryIdForView(s.ViewNumber)

	for {
		if s.viewChangeTimer == nil {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if s.id == primaryIdx {
		log.Printf("Primary %s received client request, starting view change timer for txn %d", s.ServerID, req.Transaction.Index)
		// Start a timer to trigger view change if request is not executed in time
		s.startViewChangeTimer()
		// Ensure timer stops when request is executed
		defer s.resetViewChangeTimer()
		replyMessage, err := s.IniitiatePBFT(ctx, req)
		return replyMessage, err
	} else {
		// Forward request to primary if backup
		peerId := AllServers[primaryIdx]
		peer := s.peerServers[peerId]
		if LiveSvrMap[PeerIdMap[peerId]] {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
			defer cancel() // Ensure resources are freed when the operation completes

			resp, err := peer.ClientRequest(ctx, req)
			return resp, err
		} else {
			return nil, fmt.Errorf("server is not live %s %v", PeerIdMap[peerId], LiveSvrMap[PeerIdMap[peerId]])
		}
	}
}

func (s *PBFTSvr) ValidatePrePrepareMsg(primary string, prePrepMsgData *pb.PrePrepareMessage) (bool, error) {
	sign := prePrepMsgData.LeaderSignature
	prePrepMsgData.LeaderSignature = ""
	log.Printf("to verify :: %v ", prePrepMsgData)
	isValid, err := VerifyData(prePrepMsgData, sign, s.PubKeys[primary])
	if !isValid || err != nil {
		msg := fmt.Sprintf("PrePrepare rejected due to incorrect message %v", err)
		return false, fmt.Errorf(msg)
	}
	prePrepMsgData.LeaderSignature = sign
	return true, nil
}

func (s *PBFTSvr) CreatePrepareMessage(msg *pb.PrePrepareMessage, partialSig []byte) *pb.PrepareMessage {
	// digest, _ := HashStruct(msg.Request)
	prepareMessage := &pb.PrepareMessage{
		ViewNumber:              s.ViewNumber,
		SequenceNumber:          msg.SequenceNumber,
		ReplicaId:               s.ServerID,
		ReplicaPartialSignature: partialSig,
		ClientRequestHash:       msg.ClientRequestHash,
	}
	return prepareMessage
}
