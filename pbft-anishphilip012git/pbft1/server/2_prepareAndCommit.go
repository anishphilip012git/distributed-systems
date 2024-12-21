package server

import (
	"context"
	"fmt"
	"log"
	pb "pbft1/proto"
	. "pbft1/security"
	"sync"
	"time"
)

func (s *PBFTSvr) CreateAggPrepareMessage(clientRequest *pb.ClientRequestMessage, aggregateSignature []byte, seqNo int64) *pb.PrepareReqMessage {

	log.Printf("CreateAggPrepareMessage %d", seqNo)

	commitMessage := &pb.PrepareReqMessage{
		ViewNumber: s.ViewNumber,
		PrepareCertificate: &pb.PrepareCertificate{
			ViewNumber:      s.ViewNumber,
			PrepareMessages: s.PrepareLogs[seqNo], // Quorum prepare messages
			SequenceNumber:  seqNo,
		},
		AggregatedSignature: aggregateSignature,
		// Request:             clientRequest,
		LeaderSignature: "",
		SequenceNumber:  seqNo,
	}
	sign, _ := SignData(commitMessage, s.PrivateKey)
	commitMessage.LeaderSignature = sign
	commitMessage.Request = clientRequest
	return commitMessage
}

// PreparePhaseCall: Sends the aggregated signature to replicas for final commit
func (s *PBFTSvr) PreparePhaseCall(clientRequest *pb.ClientRequestMessage, aggregateSignature []byte, seqNo int64) (int, error) {
	log.SetPrefix("PreparePhaseCall::")
	// s.Mu.Lock()
	// defer s.Mu.Unlock()
	var mu sync.Mutex
	prepareCount := 0
	aggPrepareMsg := s.CreateAggPrepareMessage(clientRequest, aggregateSignature, seqNo)
	// log.Printf("PreparePhaseCall  CreateAggPrepareMessage%v", aggPrepareMsg)
	idx := clientRequest.Transaction.Index

	// Increment commit count for self and log it
	mu.Lock()
	prepareCount++
	s.TxnState[idx] = P_STATE
	s.TxnStateBySeq[seqNo] = P_STATE
	s.AppendToCommitLogs(seqNo, aggPrepareMsg)
	log.Printf("Prepare successful for self on %s with seqNo %d -%d", s.ServerID, seqNo, aggPrepareMsg.SequenceNumber)
	mu.Unlock()

	var wg sync.WaitGroup
	// Loop through all peers to send Commit requests
	for peerId, peer := range s.peerServers {
		if peerId == PeerAddresses[s.ServerID] {
			continue
		}

		// Only proceed if the peer is live and not Byzantine
		if LiveSvrMap[PeerIdMap[peerId]] && !ByzSvrMap[PeerIdMap[peerId]] {
			wg.Add(1)
			go func(peerId string, peer pb.PBFTServiceClient) {
				defer wg.Done()

				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
				defer cancel()

				resp, err := peer.PrepareWithAggregatedSignature(ctx, aggPrepareMsg)
				if err != nil {
					log.Printf("Prepare failed for %s : %v", PeerIdMap[peerId], err)
					return
				}

				// Increment commit count for successful responses
				if resp.Success {
					mu.Lock()
					prepareCount++
					mu.Unlock()
					log.Printf("Prepare successful for %s :: response- %v", PeerIdMap[peerId], resp)
				}
			}(peerId, peer)
		}
	}

	wg.Wait()
	log.Printf("Prepare count for %s : %d / %d -> %d", s.ServerID, prepareCount, len(PeerAddresses), (len(PeerAddresses) / 2))

	// Check if majority commits were received
	if prepareCount < 2*F+1 {
		return prepareCount, fmt.Errorf("Prepare phase failed")
	}

	return prepareCount, nil
}

// CommitWithAggregatedSignature handles the RPC call for committing with an aggregated signature
func (s *PBFTSvr) PrepareWithAggregatedSignature(ctx context.Context, aggPrepareMessage *pb.PrepareReqMessage) (*pb.StatusResponse, error) {
	// Verify the commit message, e.g., check view number, validate signature, etc.
	// log.SetPrefix(fmt.Sprintf("CommitWithAggregatedSignature on %s :: ", s.ServerID))

	if !s.VerifyAggPrepareMessage(aggPrepareMessage) {
		return &pb.StatusResponse{Success: false}, fmt.Errorf("invalid commit message")
	}

	// Log the received commit message
	// log.Printf("CommitWithAggregatedSignature %s:: Received commit message from view%d: %+v", s.ServerID, aggPrepareMessage.ViewNumber, aggPrepareMessage)

	// Process the commit (update state, execute transactions, etc.)
	s.Commit(ctx, aggPrepareMessage)

	return &pb.StatusResponse{Success: true}, nil
}
func (s *PBFTSvr) Commit(ctx context.Context, msg *pb.PrepareReqMessage) error {

	if len(msg.PrepareCertificate.PrepareMessages) < 2*F+1 {
		errmsg := fmt.Sprintf("Commit rejected : len of prepare msg less than quorum gfor %d", msg.SequenceNumber)
		return fmt.Errorf(errmsg)

	}
	// log.Printf("Commit:: %s %v", s.ServerID, s.Mu.TryLock())
	s.TxnStateMu.Lock()
	defer s.TxnStateMu.Unlock()
	log.Printf("Commit call %d", msg.SequenceNumber)
	s.AppendToCommitLogs(msg.SequenceNumber, msg)
	// s.AppendCommitedTxn(msg.Request.Transaction)
	idx := msg.Request.Transaction.Index
	s.TxnState[idx] = P_STATE
	s.TxnStateBySeq[msg.SequenceNumber] = P_STATE
	return nil
}
func (s *PBFTSvr) AppendToCommitLogs(sequenceNumber int64, msg *pb.PrepareReqMessage) {
	if s.CommitedLogs == nil {
		s.CommitedLogs = map[int64]*pb.PrepareReqMessage{}
	}
	// Check if there is an existing slice for the given sequenceNumber
	if s.CommitedLogs[sequenceNumber] == nil {
		// Initialize the slice if it does not exist
		s.CommitedLogs[sequenceNumber] = &pb.PrepareReqMessage{}
	}

	// Append the message to the slice
	s.CommitedLogs[sequenceNumber] = msg //append(s.CommitedLogs[sequenceNumber], msg)
}
func (s *PBFTSvr) AppendToBufferLogs(msg *pb.Transaction, seqNo int64) {
	// Check if there is an existing slice for the given sequenceNumber
	if s.Buffer == nil {
		// Initialize the slice if it does not exist
		s.Buffer = map[int64]*pb.Transaction{}
	}

	// Append the message to the slice
	s.Buffer[seqNo] = msg
	log.Printf("AppendToBufferLogs ")
}
