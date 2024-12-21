package server

import (
	"context"
	"fmt"
	"log"
	pb "pbft1/proto"
	. "pbft1/security"
	"sync"
	"time"

	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"
)

var DigestMap = map[string]*pb.ClientRequestMessage{}

// PrePrepare Phase: Leader signs with individual signature
func (s *PBFTSvr) CreatePrePrepareMessage(clientRequest *pb.ClientRequestMessage, digest string, seqNo int64) *pb.PrePrepareMessage {
	log.Printf("CreatePrePrepareMessage request %v", clientRequest)
	prePrepareMessage := &pb.PrePrepareMessage{
		ViewNumber:        s.ViewNumber,
		SequenceNumber:    seqNo,
		Request:           clientRequest,
		LeaderSignature:   "", // Individual signature
		ClientRequestHash: digest,
	}
	log.Printf("CreatePrePrepareMessage to sign :: %v ", prePrepareMessage)
	sign, _ := SignData(prePrepareMessage, s.PrivateKey)
	prePrepareMessage.LeaderSignature = sign

	return prePrepareMessage
}

func (s *PBFTSvr) PrePreparePhaseCall(clientRequest *pb.ClientRequestMessage, seqNo int64) ([]byte, int, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	log.SetPrefix("PrePreparePhaseCall :: ")
	promiseCount := 0
	digest, _ := HashStruct(clientRequest)
	log.Printf("digest 1 :: %s -> %v", digest, clientRequest)
	DigestMap[digest] = clientRequest

	aggregateNoncePub, aggregateNonceShares, err := GenerateNonces(s.TSconfig.Config, s.TSconfig.Pub, []byte(digest), TSReplicaConfig)
	if err != nil {
		fmt.Printf("Error generating nonces: %v\n", err)
	}
	idx := clientRequest.Transaction.Index
	s.TxnStateMu.Lock()
	MsgWiseNonceQ[clientRequest.Transaction.Index] = SharedNonce{
		AggregateNoncePub: aggregateNoncePub, AggregateNonceShares: aggregateNonceShares,
	}
	s.TxnStateMu.Unlock()
	prePrepMsgData := s.CreatePrePrepareMessage(clientRequest, digest, seqNo)
	partialSigs := make([]*ted25519.PartialSignature, 0, ReplicaCount)

	// Channels to collect results from goroutines
	promiseCh := make(chan *ted25519.PartialSignature, len(s.peerServers))
	errorCh := make(chan error, len(s.peerServers))
	var wg sync.WaitGroup

	// Loop through all peers and send Prepare requests concurrently
	for peerId, peer := range s.peerServers {
		peerId := peerId
		peer := peer

		if peerId == PeerAddresses[s.ServerID] {
			// Handle self case separately
			promiseCount++
			log.Printf("on %s Promise for self debug 1", s.ServerID)
			s.AppendToPrePrepareLogs(prePrepMsgData, seqNo)
			s.TxnState[idx] = PP_STATE
			s.TxnStateBySeq[seqNo] = PP_STATE

			if !ByzSvrMap[s.ServerID] {
				ps := s.TSconfig.SignMessage([]byte(digest), MsgWiseNonceQ[idx].AggregateNonceShares[s.id], MsgWiseNonceQ[idx].AggregateNoncePub)
				partialSigs = append(partialSigs, ps)
				var partialSig []byte
				partialSig = append(partialSig, ps.ShareIdentifier)
				partialSig = append(partialSig, ps.Sig...)
				prepMsgData := s.CreatePrepareMessage(prePrepMsgData, partialSig)
				s.AppendToPrepareLogs(prepMsgData, seqNo)
				log.Printf("Promise successful for self on %s", s.ServerID)
			} else {
				log.Printf("Promise rejected for self (byz case) on %s", s.ServerID)
			}
			continue
		}

		// For other peers, launch a goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			if LiveSvrMap[PeerIdMap[peerId]] && !ByzSvrMap[PeerIdMap[peerId]] {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
				defer cancel()

				resp, err := peer.PrePrepareRequest(ctx, prePrepMsgData)
				if err != nil {
					log.Printf("Promise failed for %s : %v", s.ServerID, err)
					errorCh <- err
					return
				}

				ps := &ted25519.PartialSignature{
					ShareIdentifier: resp.ReplicaPartialSignature[0],  // first byte is ShareIdentifier
					Sig:             resp.ReplicaPartialSignature[1:], // remaining bytes are Sig
				}
				promiseCh <- ps

				log.Printf("Promise successful from %s :: response- %v", PeerIdMap[peerId], resp)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(promiseCh)
	close(errorCh)

	// Collect results from channels
	for ps := range promiseCh {
		var partialSig []byte
		partialSig = append(partialSig, ps.ShareIdentifier)
		partialSig = append(partialSig, ps.Sig...)
		prepMsgData := s.CreatePrepareMessage(prePrepMsgData, partialSig)
		s.AppendToPrepareLogs(prepMsgData, seqNo)
		partialSigs = append(partialSigs, ps)
		promiseCount++
	}

	if len(errorCh) > 0 {
		return nil, promiseCount, fmt.Errorf("PrePrepare phase failed due to errors in peer requests")
	}

	log.Printf("Promise count for %s : %d / %d -> %d", s.ServerID, promiseCount, len(PeerAddresses), (len(PeerAddresses) / 2))

	// Check if majority promises were received
	if promiseCount < 2*F+1 {
		return nil, promiseCount, fmt.Errorf("PrePrepare phase failed")
	}

	aggregateSignature := []byte{}
	valid := false
	if !ByzSvrMap[s.ServerID] {
		valid, aggregateSignature = s.AggregateAndVerify(partialSigs, []byte(digest))
		if !valid {
			return nil, promiseCount, fmt.Errorf("PrePrepare phase failed :: AggregateAndVerify step failure")
		} else {
			log.Println("AggregateAndVerify Success")
		}
	}
	return aggregateSignature, promiseCount, nil
}

// func (s *PBFTSvr) GetPartialSignature([])

// from the leader to replicas
func (s *PBFTSvr) PrePrepareRequest(ctx context.Context, msg *pb.PrePrepareMessage) (*pb.PrepareMessage, error) {
	log.Printf("PrePrepareRequest :: Replica %s received client request, starting view change timer for txn %d", s.ServerID, msg.Request.Transaction.Index)
	// Start a timer to trigger view change if request is not executed in time
	s.startViewChangeTimer()
	// Ensure timer stops when request is executed
	// log.Printf("PrePrepareRequest:: lock %v", s.Mu.TryLock())

	primary := AllServers[s.ViewNumber%ReplicaCount]
	log.Printf("----------->PrePrepareRequest from %s on %s", primary, s.ServerID)
	isValid, err := s.ValidatePrePrepareMsg(primary, msg)
	if err != nil {
		log.Printf("ValidatePrePrepareMsg  Error %v", err)
	}
	errMsg := ""
	if !isValid || err != nil { // Byz replica case added , in prepare case
		errMsg = fmt.Sprintf("PrePrepare rejected %v", err)
	}
	if msg.ViewNumber != s.ViewNumber {
		errMsg = fmt.Sprintf("Prepare rjected due to incorrect view number svr has %d, msg has %d", s.ViewNumber, msg.ViewNumber)
	}
	if msg.SequenceNumber < s.stableCheckpoint || msg.SequenceNumber > (s.stableCheckpoint+s.waterMarkRange) {
		errMsg = fmt.Sprintf("Prepare rjected due to incorrect sequence number: svr has %d (Lasstablechkpoint : %d Watermark %d ), msg has %d", s.SequenceNumber, s.stableCheckpoint, s.waterMarkRange, msg.SequenceNumber)
	}

	// // check if record has alread a msg with same seq no
	// if _, ok := s.prePrepareLogs[msg.SequenceNumber]; ok {
	// 	errMsg = fmt.Sprintf("Prepare rjected due to incorrect SEQUENCE number svr has %d, msg has %d", s.sequenceNumber, msg.SequenceNumber)
	// }
	if errMsg != "" {
		return s.CreatePrepareMessage(msg, nil), fmt.Errorf(errMsg)
	}

	// Record the PrePrepare message
	s.Mu.Lock()
	s.AppendToPrePrepareLogs(msg, msg.SequenceNumber)
	idx := msg.Request.Transaction.Index
	s.TxnState[idx] = PP_STATE
	s.SequenceNumber = msg.SequenceNumber
	seqNo := msg.SequenceNumber
	s.TxnStateBySeq[seqNo] = PP_STATE
	s.Mu.Unlock()
	if ByzSvrMap[s.ServerID] {
		errMsg = fmt.Sprintf("PrePrepare rejected by repliaca (byz replica)")
		return s.CreatePrepareMessage(msg, nil), fmt.Errorf(errMsg)
	}

	// threshold signing by replicas
	digest, _ := HashStruct(msg.Request)
	ps := s.TSconfig.SignMessage([]byte(digest), MsgWiseNonceQ[idx].AggregateNonceShares[s.id], MsgWiseNonceQ[idx].AggregateNoncePub)
	var partialSig []byte
	partialSig = append(partialSig, ps.ShareIdentifier) // append ShareIdentifier
	partialSig = append(partialSig, ps.Sig...)          // append Sig bytes
	return s.CreatePrepareMessage(msg, partialSig), nil
}
