package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	pb "tpcbyz/proto"
	. "tpcbyz/security"
)

func (s *TPCServer) ValidatePrepareMsg(primary string, prepMsgData *pb.PrepareRequest) (bool, error) {
	sign := prepMsgData.LeaderSignature
	prepMsgData.LeaderSignature = ""
	// log.Printf("ValidatePrepareMsg to verify :: %v ", prepMsgData)
	isValid, err := VerifyData(prepMsgData, sign, s.PubKeys[primary])
	if !isValid || err != nil {
		msg := fmt.Sprintf("PrePrepare rejected due to incorrect message %v", err)
		return false, fmt.Errorf(msg)
	}
	isValid, err = s.ValidatePreparCertificate(prepMsgData)
	// prepMsgData.LeaderSignature = sign
	return isValid, err
}

func (s *TPCServer) ValidatePreparCertificate(prepMsgData *pb.PrepareRequest) (bool, error) {

	validCnt := 0
	for nodeid, signStr := range prepMsgData.Signatures {
		isValid, _ := VerifyData(prepMsgData.Digest, signStr, s.PubKeys[nodeid])
		if isValid {
			validCnt++
		} else {
			log.Printf("ValidatePreparCertificate::Invalid signature in prepare certificate for %s", nodeid)
		}
	}

	if validCnt >= 2*F+1 {
		return true, nil
	} else {
		msg := fmt.Sprintf("Not enough Valid Certificates: %d", validCnt)
		log.Println(msg)
		return false, fmt.Errorf(msg)
	}

}
func (s *TPCServer) CreateCommitRequestCertificate(msg *pb.PrepareRequest, signatureArr []*pb.SignedMessage) *pb.CommitRequest {
	// digest, _ := HashStruct(msg.Request)
	signedSigMap := map[string]string{}

	for i, _ := range signatureArr {
		signedSigMap[signatureArr[i].NodeId] = signatureArr[i].Signature
	}
	commitMsg := &pb.CommitRequest{
		ViewNumber:     msg.ViewNumber,
		SequenceNumber: msg.SequenceNumber,
		Digest:         msg.Digest,
		Signatures:     signedSigMap,
	}
	// log.Printf("CreateCommitRequestCertificate to sign :: %v ", commitMsg)
	sign, _ := SignData(commitMsg, s.PrivateKey)
	commitMsg.LeaderSignature = sign

	return commitMsg
}

// AppendToPrepareLogs records prepare messages for a specific sequence number
func (s *TPCServer) AppendToPrepareLogs(prepareMessage *pb.PrepareRequest, seqNo int64) {
	if s.PrepareLogs == nil {
		s.PrepareLogs = map[int64]*pb.PrepareRequest{}
	}
	s.PrepareLogs[seqNo] = prepareMessage
}
func (s *TPCServer) CreateCommitMessage(msg *pb.PrepareRequest, sign string) *pb.CommitMessage {
	signedmsg := &pb.SignedMessage{
		NodeId:    s.ServerID,
		Signature: sign,
	}
	// digest, _ := HashStruct(msg.Request)
	commitMsg := &pb.CommitMessage{
		ViewNumber:        s.ViewNumber,
		SequenceNumber:    msg.SequenceNumber,
		ClientRequestHash: msg.Digest,
		SignedMsg:         signedmsg,
	}

	return commitMsg
}

// PreparePhaseCall sends prepare messages to replicas
func (s *TPCServer) PreparePhaseCall(clientRequest *pb.ClientRequestMessage, prepareRequest *pb.PrepareRequest, seqNo int64) (*pb.CommitRequest, int, error) {
	prefix := ("Prepare PhaseCall::")
	// var mu sync.Mutex
	prepareCount := 0

	responseCh := make(chan *pb.CommitMessage, len(s.peerServers))
	errorCh := make(chan error, len(s.peerServers))
	signatureArr := []*pb.SignedMessage{}
	var wg sync.WaitGroup
	for peerId, peer := range s.peerServers {
		peerId := peerId
		peer := peer
		if ServerClusters[PeerIdMap[peerId]] != ServerClusters[s.ServerID] {
			continue
		}
		if peerId == PeerAddresses[s.ServerID] {
			prepareCount++
			log.Printf(prefix+"on %s Promise for self debug 1", s.ServerID)
			// mu.Lock()
			s.AppendToPrepareLogs(prepareRequest, seqNo)
			s.TxnState[clientRequest.Transaction.Index] = P_STATE
			s.TxnStateBySeq[seqNo] = P_STATE
			// mu.Unlock()
			sign, _ := SignData(prepareRequest.Digest, s.PrivateKey)
			commitMsg := s.CreateCommitMessage(prepareRequest, sign)
			signatureArr = append(signatureArr, commitMsg.SignedMsg)
			// s.AppendToPrepareLogs(prepMsgData, seqNo)
			log.Printf(prefix+"Promise successful for self on %s", s.ServerID)

			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if LiveSvrMap[PeerIdMap[peerId]] && !ByzSvrMap[PeerIdMap[peerId]] {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(rpcCtxDeadlineTimeout))
				defer cancel()

				resp, err := peer.PBFTPrepare(ctx, prepareRequest)
				if err != nil {
					log.Printf(prefix+"Promise failed for %s : %v", s.ServerID, err)
					errorCh <- err
					return
				}

				responseCh <- resp
				log.Printf(prefix+"Promise successful from %s :: response- %v", PeerIdMap[peerId], resp)
			}
		}()
	}

	wg.Wait()
	close(responseCh)
	close(errorCh)

	for resp := range responseCh {
		if resp != nil {
			isValid, _ := VerifyData(prepareRequest.Digest, resp.SignedMsg.Signature, s.PubKeys[resp.SignedMsg.NodeId])
			if isValid {
				signatureArr = append(signatureArr, resp.SignedMsg)
				prepareCount++
			}
		}
	}
	log.Printf(prefix+"Prepare count for %s : %d / %d -> %d", s.ServerID, prepareCount, s.ReplicaCount, 2*F+1)

	if prepareCount < 2*F+1 {
		return nil, prepareCount, fmt.Errorf("Prepare phase failed")
	}
	commitreq := s.CreateCommitRequestCertificate(prepareRequest, signatureArr)
	return commitreq, prepareCount, nil
}

// PBFTPrepare handles incoming prepare messages
func (s *TPCServer) PBFTPrepare(ctx context.Context, msg *pb.PrepareRequest) (*pb.CommitMessage, error) {
	log.Printf("PBFTPrepare :: Received prepare message for seqNo %d", msg.SequenceNumber)
	primary := ClusterMapping[ServerClusters[s.ServerID]].Leader
	isValid, err := s.ValidatePrepareMsg(primary, msg)
	if err != nil {
		log.Printf("ValidatePrePrepareMsg  Error %v", err)
	}
	if !isValid {
		return nil, fmt.Errorf("Invalid Prepare message")
	}
	s.AppendToPrepareLogs(msg, msg.SequenceNumber)
	var idx int64
	clientreq, exists := DigestMap[msg.Digest]
	if exists {
		idx = clientreq.Transaction.Index
	} else {
		return nil, fmt.Errorf("Invalid Prepare message")
	}
	s.Mu.Lock()
	s.TxnState[idx] = P_STATE
	s.TxnStateBySeq[msg.SequenceNumber] = P_STATE
	s.Mu.Unlock()
	sign, _ := SignData(msg.Digest, s.PrivateKey)
	commitMsg := s.CreateCommitMessage(msg, sign)
	return commitMsg, nil
}
