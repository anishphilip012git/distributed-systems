package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	pbft "pbft1/proto"
	. "pbft1/security"
	"sync"
	"time"

	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"
)

// CreateCheckpoint initiates a checkpoint and broadcasts it to other replicas.
func (replica *PBFTSvr) CreateCheckpoint(sequenceNumber int64) {
	replica.Mu.Lock()
	defer replica.Mu.Unlock()
	log.Printf("CreateCheckpoint :: for sequenceNumber %d", sequenceNumber)
	stateDigest := generateStateDigest(replica.getState()) // getState retrieves current replica state

	checkpointMsg := &pbft.CheckpointMessage{
		SequenceNumber: sequenceNumber,
		StateDigest:    stateDigest,
		ReplicaId:      replica.ServerID,
		ViewNumber:     replica.ViewNumber,
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
	defer cancel()
	success, aggregateSign, err := replica.BroadcastCheckpoint(ctx, checkpointMsg)
	if err != nil {
		return
	}
	log.Printf("CreateCheckpoint :: Stage 1: %v", success)
	checkpointMsg.AggregatedSignature = aggregateSign
	success, err = replica.BroadcastCheckpoint2(ctx, checkpointMsg)
	if err != nil {
		return
	}
	log.Printf("CreateCheckpoint :: Stage 2: %v", success)
}

func (s *PBFTSvr) CreateAggregatedCheckpointMsg(partialSig []byte) *pbft.ReplicaChkptMsg {

	checkpointMsg := &pbft.ReplicaChkptMsg{
		ReplicaId:               s.ServerID,
		ReplicaPartialSignature: partialSig,
	}
	return checkpointMsg

}

// BroadcastCheckpoint sends checkpoint messages to all peers.
func (s *PBFTSvr) BroadcastCheckpoint(ctx context.Context, checkpointMsg *pbft.CheckpointMessage) (bool, []byte, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	for i := int(s.stableCheckpoint); i < int(checkpointMsg.SequenceNumber); i++ {
		if val, ok := s.TxnStateBySeq[checkpointMsg.SequenceNumber]; !ok || val != E_STATE {
			log.Printf("Missing Transactions :: seq %d:: with state: %s can't be checkpointed", i, val)
			return false, nil, fmt.Errorf("Missing Transactions :: can't be checkpointed")
		}
	}

	// Step 1: Generate the digest of the checkpoint message
	digest, err := HashStruct(checkpointMsg)
	if err != nil {
		log.Printf("Error generating digest: %v", err)
		return false, nil, err
	}

	// Step 2: Generate nonces
	aggregateNoncePub, aggregateNonceShares, err := GenerateNonces(s.TSconfig.Config, s.TSconfig.Pub, []byte(digest), TSReplicaConfig)
	if err != nil {
		log.Printf("Error generating nonces: %v", err)
		return false, nil, err
	}

	// Store nonce data in a map for future reference
	idx := checkpointMsg.ViewNumber*100 + checkpointMsg.SequenceNumber
	ChkpntMsgNonceQ[idx] = SharedNonce{
		AggregateNoncePub:    aggregateNoncePub,
		AggregateNonceShares: aggregateNonceShares,
	}

	// Step 3: Collect partial signatures
	var mu sync.Mutex
	partialSigs := make([]*ted25519.PartialSignature, 0, ReplicaCount)

	// Add self-signature
	ps := s.TSconfig.SignMessage([]byte(digest), ChkpntMsgNonceQ[idx].AggregateNonceShares[s.id], ChkpntMsgNonceQ[idx].AggregateNoncePub)
	mu.Lock()
	partialSigs = append(partialSigs, ps)
	mu.Unlock()
	log.Printf("Partial signature collected from self on server %s", s.ServerID)
	// Step 4: Broadcast checkpoint message to peers and collect their partial signatures
	var wg sync.WaitGroup
	for peerId, peer := range s.peerServers {
		if peerId == s.ServerID {
			continue // Skip self
		}
		wg.Add(1)
		go func(peerId string, p pbft.PBFTServiceClient) {
			defer wg.Done()

			// Send the checkpoint message to the peer
			resp, err := s.ProcessCheckpoint(ctx, checkpointMsg)
			if err != nil {
				log.Printf("Failed to process checkpoint with peer %s: %v", peerId, err)
				return
			}
			// Collect the partial signature from the peer
			ps := ted25519.PartialSignature{
				ShareIdentifier: resp.ReplicaPartialSignature[0],  // first byte is ShareIdentifier
				Sig:             resp.ReplicaPartialSignature[1:], // remaining bytes are Sig
			}
			// Append the partial signature to the list
			mu.Lock()
			partialSigs = append(partialSigs, &ps)
			mu.Unlock()
		}(peerId, peer)
	}

	// Wait for all peers to respond
	wg.Wait()

	// Step 5: Check if we have enough partial signatures to meet the threshold
	if len(partialSigs) < F+1 { // assuming `f+1` is the threshold for quorum
		log.Printf("Failed to collect enough partial signatures for checkpoint %d", checkpointMsg.SequenceNumber)
		return false, nil, fmt.Errorf("insufficient partial signatures")
	}

	// Step 6: Combine partial signatures into an aggregated signature
	aggregateSignature := []byte{}
	valid := false
	if !ByzSvrMap[s.ServerID] { // skipping agg verificcation and preventing prepare if primary is byzantine

		valid, aggregateSignature = s.AggregateAndVerify(partialSigs, []byte(digest))
		if !valid {
			return false, nil, fmt.Errorf("PrePrepare phase failed :: AggregateAndVerify step failure")
		} else {
			log.Println("AggregateAndVerify Success")
		}
	}
	// Successfully collected quorum of partial signatures
	return true, aggregateSignature, nil
}

// BroadcastCheckpoint sends checkpoint messages to all peers.
func (s *PBFTSvr) BroadcastCheckpoint2(ctx context.Context, checkpointMsg *pbft.CheckpointMessage) (bool, error) {
	updated := make([]bool, 0, ReplicaCount)
	// Step 4: Broadcast checkpoint message to peers and collect their partial signatures
	var wg sync.WaitGroup
	for peerId, peer := range s.peerServers {
		if peerId == s.ServerID {
			continue // Skip self
		}
		wg.Add(1)
		go func(peerId string, p pbft.PBFTServiceClient) {
			defer wg.Done()

			// Send the checkpoint message to the peer
			resp, err := s.ProcessCheckpoint(ctx, checkpointMsg)
			if err != nil {
				log.Printf("Failed to process checkpoint with peer %s: %v", peerId, err)
				return
			}
			if resp == nil {
				mu.Lock()
				updated = append(updated, true)
				mu.Unlock()
			}
		}(peerId, peer)
	}

	// Wait for all peers to respond
	wg.Wait()

	// Step 5: Check if we have enough partial signatures to meet the threshold
	if len(updated) < F+1 { // assuming `f+1` is the threshold for quorum
		log.Printf("Failed to replicate partial aggregated signs for checkpoint %d", checkpointMsg.SequenceNumber)
		return false, fmt.Errorf("insufficient aggreagted signatures")
	}

	// Successfully collected quorum of partial signatures
	return true, nil
}

// ProcessCheckpoint handles a received checkpoint message from a peer.
func (s *PBFTSvr) ProcessCheckpoint(ctctx context.Context, msg *pbft.CheckpointMessage) (*pbft.ReplicaChkptMsg, error) {
	// Log the checkpoint and check if a stable quorum is achieved.
	// replica.CheckpointLog[msg.SequenceNumber] = append(replica.CheckpointLog[msg.SequenceNumber], msg)
	// if len(replica.CheckpointLog[msg.SequenceNumber]) >= 2*replica.f+1 {
	// 	replica.stabilizeCheckpoint(msg.SequenceNumber)
	// }
	if msg.AggregatedSignature != nil {
		s.Checkpoints[msg.SequenceNumber] = msg
		s.stabilizeCheckpoint(msg.SequenceNumber)
		log.Printf("Updated Replica with check[oint signature")
		return nil, nil
	}
	idx := msg.ViewNumber*100 + msg.SequenceNumber
	digest, _ := HashStruct(msg)
	ps := s.TSconfig.SignMessage([]byte(digest), ChkpntMsgNonceQ[idx].AggregateNonceShares[s.id], ChkpntMsgNonceQ[idx].AggregateNoncePub)
	var partialSig []byte
	partialSig = append(partialSig, ps.ShareIdentifier) // append ShareIdentifier
	partialSig = append(partialSig, ps.Sig...)
	checkpointMsg := &pbft.ReplicaChkptMsg{
		ReplicaId:               s.ServerID,
		ReplicaPartialSignature: partialSig,
	}
	return checkpointMsg, nil
}

// stabilizeCheckpoint stabilizes a checkpoint and prunes older logs.
func (replica *PBFTSvr) stabilizeCheckpoint(sequenceNumber int64) {
	replica.stableCheckpoint = sequenceNumber
	replica.pruneLogs(sequenceNumber)
}

// pruneLogs removes logs with sequence numbers below the stable checkpoint.
func (replica *PBFTSvr) pruneLogs(sequenceNumber int64) {
	// Clear logs below the stable checkpoint
	for seq := range replica.PrePrepareLogs {
		if seq <= sequenceNumber {
			delete(replica.PrePrepareLogs, seq)
		}
	}
	for seq := range replica.PrepareLogs {
		if seq <= sequenceNumber {
			delete(replica.PrepareLogs, seq)
		}
	}
	// Repeat for other log types as necessary...
}

type PartialState struct {
	BalanceMap    map[string]int   `json:"balance_map"`
	TxnStateBySeq map[int]string   `json:"txn_state_by_seq"`
	TxnState      map[int64]string `json:"txn_state"`
}

// Helper to get the current state (example function, depends on your application state)
func (replica *PBFTSvr) getState() []byte {
	// Your logic to retrieve state data to hash for the digest
	stateMap := map[string]interface{}{
		"balance_map":      replica.BalanceMap,
		"txn_state_by_seq": replica.TxnStateBySeq,
		"txn_state":        replica.TxnState,
	}

	stateJson, err := json.Marshal(stateMap)
	if err != nil {
		log.Printf("getState::Marshal error %v", err)
		return nil
	}
	return []byte(stateJson)
}

// Helper function to generate a state digest from state data
func generateStateDigest(state []byte) string {
	hash := sha256.Sum256(state)
	return hex.EncodeToString(hash[:])
}
