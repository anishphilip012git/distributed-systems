package server

import (
	"context"
	"log"
	"math"
	pb "pbft1/proto"

	// . "pbft1/security"
	"sync/atomic"
	"time"
)

// const (
// 	viewChangeTimeout   = 10 * time.Second // Timeout for triggering a view change
// 	rpcCtxDeadlineTimeout = 5 * time.Second // Deadline for RPC communication
// 	checkpointInterval  = 10               // Interval for stable checkpointing
// 	waterMarkBuffer     = 5                // Buffer for watermarks
// 	ReplicaCount        = 4                // Total replicas in the system
// 	PREPARED            = "PREPARED"       // State indicating a message is prepared
// )

// // PBFT Server struct
// type PBFTSvr struct {
// 	ServerID             string
// 	ViewNumber           int64
// 	stableCheckpoint     int64
// 	lowWaterMark         int64
// 	highWaterMark        int64
// 	Log                  map[int64]*pb.PreparedMessage
// 	Checkpoints          map[int64][]*pb.CheckpointMessage
// 	viewChangeTimer      *time.Timer
// 	viewChangeResponses  map[int64][]*pb.ViewChangeMessage
// 	peerServers          map[string]pb.PBFTServiceClient
// 	lastSequenceNumber   int64
// 	Mu                   sync.Mutex
// 	executionComplete    chan struct{}
// }

// // Start a timer to trigger a view change
// func (s *PBFTSvr) startViewChangeTimer() {
// 	if s.viewChangeTimer == nil {
// 		log.Printf("View change timer started for view %d at replica %s", s.ViewNumber, s.ServerID)
// 		s.viewChangeTimer = time.AfterFunc(viewChangeTimeout, func() {
// 			log.Printf("View change timer expired for view %d at replica %s", s.ViewNumber, s.ServerID)
// 			// s.Mu.Unlock()
// 			 s.TriggerViewChange()
// 		})
// 	}
// }
// // Reset the view change timer
// func (s *PBFTSvr) resetViewChangeTimer() {
// 	if s.viewChangeTimer != nil {
// 		s.viewChangeTimer.Stop()
// 		s.viewChangeTimer = nil
// 	}
// }

// Start a timer to trigger a view change if the request is not executed in time.
func (s *PBFTSvr) startViewChangeTimer() {
	st := time.Now()
	// Start only if no view change is in progress and timer isn't already set
	if atomic.LoadInt32(&s.viewChangeInProgress) == 0 && s.viewChangeTimer == nil {
		log.Printf("View change timer started for view %d at replica %s", s.ViewNumber, s.ServerID)
		// Use time.AfterFunc to trigger view change if timeout expires
		s.viewChangeTimer = time.AfterFunc(viewChangeTimeout, func() {
			// Attempt to set the view change in progress atomically
			if atomic.CompareAndSwapInt32(&s.viewChangeInProgress, 0, 1) {
				// Only execute if timer hasn't been reset
				// s.Mu.Lock()
				// defer s.Mu.Unlock()
				s.viewChangeMutex.Lock()
				defer s.viewChangeMutex.Unlock()
				if s.viewChangeTimer != nil {
					log.Printf("View change timer expired for view %d at replica %s", s.ViewNumber, s.ServerID)
					s.TriggerViewChange() // Initiate view change
				}
			}
		})
	}
	log.Printf("startViewChangeTimer %v", time.Since(st))
}

// Reset the view change timer, ensuring no redundant triggers.
func (s *PBFTSvr) resetViewChangeTimer() {
	if s.viewChangeTimer != nil {
		s.viewChangeTimer.Stop()
		s.viewChangeTimer = nil
		atomic.StoreInt32(&s.viewChangeInProgress, 0) // Allow future view change triggers
		log.Printf("View change timer reset at replica %s", s.ServerID)
	}
}

// Collect PREPARED messages after the latest checkpoint
func (s *PBFTSvr) getPreparedMessagesAfterCheckpoint(checkpoint int64) []*pb.PrepareCertificate {
	var preparedMessages []*pb.PrepareCertificate
	for seqNum, msg := range s.CommitedLogs {
		if seqNum > checkpoint && seqNum <= (checkpoint+s.waterMarkRange) { //&& msg == P_STATE
			preparedMessages = append(preparedMessages, msg.PrepareCertificate)
		}
	}
	return preparedMessages
}

// Get checkpoint proof for the stable checkpoint
func (s *PBFTSvr) getCheckpointProof(checkpoint int64) []*pb.CheckpointMessage {
	ans := []*pb.CheckpointMessage{}
	ans = append(ans, s.Checkpoints[checkpoint])
	return ans
}

// Trigger a view change when timeout or failure is detected
func (s *PBFTSvr) TriggerViewChange() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	newView := s.ViewNumber + 1
	lastCheckpoint := s.stableCheckpoint
	preparedCertis := s.getPreparedMessagesAfterCheckpoint(lastCheckpoint)

	// Create VIEW-CHANGE message
	viewChangeMsg := &pb.ViewChangeMessage{
		ViewNumber:           newView,
		LastSequenceNumber:   lastCheckpoint,
		CheckpointMsgs:       s.getCheckpointProof(lastCheckpoint),
		PreparedCertificates: preparedCertis,
		ReplicaId:            s.ServerID,
	}

	log.Printf("TriggerViewChange :: Replica %s initiating view change to view %d", s.ServerID, newView)

	// Broadcast VIEW-CHANGE message to all replicas
	s.broadcastViewChange(viewChangeMsg)
	log.SetPrefix("")
	// Wait for sufficient responses to enter the new view
	go s.waitForNewView(newView)
}

// Wait for sufficient VIEW-CHANGE messages to proceed with a NEW-VIEW
func (s *PBFTSvr) waitForNewView(newView int64) {
	// log.SetPrefix("waitForNewView ::")
	requiredQuorum := 2*F + 1
	timer := time.NewTimer(viewChangeTimeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.Printf("waitForNewView :: Timeout while waiting for sufficient VIEW-CHANGE messages for view %d", newView)
			return
		default:
			// s.viewChangeMutex.Lock()
			if len(s.viewChangeResponses[newView]) >= requiredQuorum {
				// s.viewChangeMutex.Unlock()
				s.enterNewView(newView)
				return
			}
			// log.Printf("waitForNewView:: waitn g for quorum")
			// s.viewChangeMutex.Unlock()
			time.Sleep(100 * time.Millisecond) // Wait briefly before checking again
		}
		log.SetPrefix("")
	}
}

// Broadcast a VIEW-CHANGE message to all replicas
func (s *PBFTSvr) broadcastViewChange(msg *pb.ViewChangeMessage) {
	for peerID, peer := range s.peerServers {
		if peerID == s.ServerID {
			// Self-processing: Record the message locally
			log.Printf("Received VIEW-CHANGE message for view %d from replica %s at %s", msg.ViewNumber, msg.ReplicaId, s.ServerID)
			log.Printf("Received VIEW-CHANGE message for view %v from replica %s", msg, s.ServerID)
			s.recordViewChangeResponse(msg)
			continue
		}

		go func(peer pb.PBFTServiceClient) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeoutViewChangeCalls))
			defer cancel()

			_, err := peer.ViewChange(ctx, msg)
			if err != nil {
				log.Printf("Failed to send VIEW-CHANGE to replica %s: %v", peerID, err)
			}
		}(peer)
	}
}

// RPC handler for receiving VIEW-CHANGE messages
func (s *PBFTSvr) ViewChange(ctx context.Context, msg *pb.ViewChangeMessage) (*pb.StatusResponse, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	// Validate and record the message
	if msg.ViewNumber <= s.ViewNumber {
		log.Printf("Ignoring outdated VIEW-CHANGE message for msg view %d and svr view %d ", msg.ViewNumber, s.ViewNumber)
		return nil, nil
	}
	log.Printf("Received VIEW-CHANGE message for view %d from replica %s at %s", msg.ViewNumber, msg.ReplicaId, s.ServerID)
	log.Printf("Received VIEW-CHANGE message for view %v from replica %s", msg, s.ServerID)
	// Validate
	// if len(msg.CheckpointMsgs) > 0 && msg.CheckpointMsgs[0].AggregatedSignature == nil {
	// 	log.Printf("AggregatedSignature not present for checkpoint VIEW-CHANGE message for view %d at %s", msg.ViewNumber, s.ServerID)
	// 	return nil, nil
	// }
	s.recordViewChangeResponse(msg)
	return &pb.StatusResponse{Success: true}, nil
}

// Record a VIEW-CHANGE response and trigger new view if quorum is reached
func (s *PBFTSvr) recordViewChangeResponse(msg *pb.ViewChangeMessage) {
	s.viewChangeResponses[msg.ViewNumber] = append(s.viewChangeResponses[msg.ViewNumber], msg)
	requiredQuorum := (2*F + 1) //F+1
	log.Printf("recordViewChangeResponse at %s , len of vc repsonses= %d", s.ServerID, len(s.viewChangeResponses[msg.ViewNumber]))
	if len(s.viewChangeResponses[msg.ViewNumber]) >= requiredQuorum {
		if s.id == s.getPrimaryIdForView(msg.ViewNumber) {
			s.sendNewViewMessage(msg.ViewNumber)
		}
	}
}
func (s *PBFTSvr) getPrimaryIdForView(newView int64) int64 {
	return newView % ReplicaCount
}

// func (s *PBFTSvr) collectViewChange(newView int64, msg *pb.ViewChangeMessage) {
// 	// s.Mu.Lock()
// 	// defer s.Mu.Unlock()

// 	// Append the message to the corresponding view's response list
// 	s.viewChangeResponses[newView] = append(s.viewChangeResponses[newView], msg)
// }

// Send NEW-VIEW message once quorum is achieved
func (s *PBFTSvr) sendNewViewMessage(newView int64) {
	if ByzSvrMap[s.ServerID] {
		return
	}
	// s.Mu.Lock()
	// defer s.Mu.Unlock()
	log.Printf("sendNewViewMessage:: at %s", s.ServerID)
	viewChanges := s.viewChangeResponses[newView]
	preparedLogs := s.computeNewPrePrepare(viewChanges)

	newViewMsg := &pb.NewViewMessage{
		ViewNumber:         newView,
		ViewChangeMessages: viewChanges,
		PrePrepareMessages: preparedLogs,
	}

	s.broadcastNewView(newViewMsg)

}

// Broadcast a VIEW-CHANGE message to all replicas
func (s *PBFTSvr) broadcastNewView(msg *pb.NewViewMessage) {

	log.Printf("broadcastNewView :: msg=%v", msg)
	for peerID, peer := range s.peerServers {
		if peerID == s.ServerID {
			// Self-processing: Record the message locally
			if msg.ViewNumber <= s.ViewNumber {
				log.Printf("Ignoring outdated NEW-VIEW message for view %d", msg.ViewNumber)
				continue
			}
			s.enterNewView(msg.ViewNumber)
			continue
		}

		go func(peer pb.PBFTServiceClient) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeoutViewChangeCalls))
			defer cancel()

			_, err := peer.NewView(ctx, msg)
			if err != nil {
				log.Printf("Failed to send NEW-VIEW to replica %s: %v", peerID, err)
			}
		}(peer)
	}
}

// RPC handler for receiving NEW-VIEW messages
func (s *PBFTSvr) NewView(ctx context.Context, msg *pb.NewViewMessage) (*pb.StatusResponse, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	log.Printf("Received NEW-VIEW message on server %s view %s", s.ServerID, msg)

	if msg.ViewNumber <= s.ViewNumber {
		log.Printf("Ignoring outdated NEW-VIEW message for view %d", msg.ViewNumber)
		return nil, nil
	}

	s.replayPrePrepareLogs(ctx, msg.PrePrepareMessages)
	s.enterNewView(msg.ViewNumber)
	return &pb.StatusResponse{Success: true}, nil
}

// Enter the new view and resume normal operations
func (s *PBFTSvr) enterNewView(viewNumber int64) {
	log.Printf("enterNewView :: Replica %s entering new view %d", s.ServerID, viewNumber)
	s.ViewNumber = viewNumber
	s.resetViewChangeTimer()
}

func (s *PBFTSvr) replayPrePrepareLogs(ctx context.Context, prePrepareSet []*pb.PrePrepareMessage) {
	log.Printf("replayPrePrepareLogs TryLock %v", s.Mu.TryLock())
	if len(prePrepareSet) > 0 {
		for _, prePrepareMsg := range prePrepareSet {
			seqNum := prePrepareMsg.SequenceNumber

			s.IniitiatePBFT(ctx, prePrepareMsg.Request)
			log.Printf("Replayed PrePrepare log for sequence number %d in view %d", seqNum, prePrepareMsg.ViewNumber)
		}
	}
}

// Compute new PRE-PREPARE messages based on VIEW-CHANGE logs
func (s *PBFTSvr) computeNewPrePrepare(viewChanges []*pb.ViewChangeMessage) []*pb.PrePrepareMessage {
	log.Printf("computeNewPrePrepare got %d vc msgs", len(viewChanges))
	var prePrepareMessages []*pb.PrePrepareMessage
	log.Printf("computeNewPrePrepare got %v vc msgs", viewChanges)
	// Find min-s as the latest stable checkpoint across all view-change messages
	minSequence := int64(math.MaxInt64)
	for _, vcMsg := range viewChanges {
		for _, checkpoint := range vcMsg.CheckpointMsgs {
			if checkpoint.SequenceNumber < minSequence {
				minSequence = checkpoint.SequenceNumber
			}
		}
	}

	// Find max-s as the highest sequence number in any PrepareCertificate across all view-change messages
	maxSequence := minSequence
	for _, vcMsg := range viewChanges {
		for _, cert := range vcMsg.PreparedCertificates {
			for _, prepareMsg := range cert.PrepareMessages {
				if prepareMsg.SequenceNumber > maxSequence {
					maxSequence = prepareMsg.SequenceNumber
				}
			}
		}
	}

	// Create new PRE-PREPARE messages for each sequence number between min-s and max-s
	for seqNum := minSequence + 1; seqNum <= maxSequence; seqNum++ {
		found := false
		var digest string

		// Check if there is at least one PrepareCertificate with this sequence number
		for _, vcMsg := range viewChanges {
			for _, cert := range vcMsg.PreparedCertificates {
				if cert.PrepareMessages[0].SequenceNumber == seqNum {
					digest = cert.PrepareMessages[0].ClientRequestHash
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// Create a new PRE-PREPARE message with the found digest or a null digest if not found
		if !found {
			digest = "dnull" // "dnull" is the digest for a special null request
		}
		s.SequenceNumber++
		prePrepareMsg := s.CreatePrePrepareMessage(DigestMap[digest], digest, s.SequenceNumber)
		// Construct the PRE-PREPARE message
		prePrepareMessages = append(prePrepareMessages, prePrepareMsg)
	}
	log.Printf("computeNewPrePrepare figuredout %d prePrepareMessages msgs", len(prePrepareMessages))
	return prePrepareMessages
}
