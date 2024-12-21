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

var DigestMap = map[string]*pb.ClientRequestMessage{}

func (s *TPCServer) ValidatePrePrepareMsg(primary string, prePrepMsgData *pb.PrePrepareMessage) (bool, error) {
	sign := prePrepMsgData.LeaderSignature
	prePrepMsgData.LeaderSignature = ""
	log.Printf("ValidatePrePrepareMsg to verify :: %v ", prePrepMsgData)
	isValid, err := VerifyData(prePrepMsgData.ClientRequestHash, sign, s.PubKeys[primary])
	if !isValid || err != nil {
		msg := fmt.Sprintf("PrePrepare rejected due to incorrect message %v", err)
		return false, fmt.Errorf(msg)
	}
	// prePrepMsgData.LeaderSignature = sign
	return true, nil
}

// PrePrepare Phase: Leader signs with individual signature
func (s *TPCServer) CreatePrePrepareMessage(clientRequest *pb.ClientRequestMessage, digest string, seqNo int64) *pb.PrePrepareMessage {
	log.Printf("CreatePrePrepareMessage request %v", clientRequest)
	prePrepareMessage := &pb.PrePrepareMessage{
		ViewNumber:        s.ViewNumber,
		SequenceNumber:    seqNo,
		Request:           clientRequest,
		LeaderSignature:   "", // Individual signature
		ClientRequestHash: digest,
	}
	log.Printf("CreatePrePrepareMessage to sign :: %v ", prePrepareMessage)
	sign, _ := SignData(digest, s.PrivateKey)
	prePrepareMessage.LeaderSignature = sign

	return prePrepareMessage
}

func (s *TPCServer) AppendToPrePrepareLogs(prePrepareMsg *pb.PrePrepareMessage, seqNo int64) {
	if s.PrePrepareLogs == nil {
		s.PrePrepareLogs = map[int64]*pb.PrePrepareMessage{}
	}
	s.PrePrepareLogs[seqNo] = prePrepareMsg
}

func (s *TPCServer) PrePreparePhaseCall(clientRequest *pb.ClientRequestMessage, seqNo int64) (*pb.PrepareRequest, int, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	prefix := ("PrePreparePhaseCall :: ")
	promiseCount := 0
	digest, _ := HashStruct(clientRequest)
	log.Printf(prefix+"digest 1 :: %s -> %v", digest, clientRequest)
	DigestMap[digest] = clientRequest

	prePrepMsgData := s.CreatePrePrepareMessage(clientRequest, digest, seqNo)

	// Channels to collect results from goroutines
	responseCh := make(chan *pb.PrepareMessage, len(s.peerServers))
	errorCh := make(chan error, len(s.peerServers))
	var wg sync.WaitGroup
	signatureArr := []*pb.SignedMessage{}
	for peerId, peer := range s.peerServers {
		peerId := peerId
		peer := peer

		if ServerClusters[PeerIdMap[peerId]] != ServerClusters[s.ServerID] {
			continue
		}
		// log.Println(prefix+"ServerCInvalid PrePrepare messagelusters[peerId] , ServerClusters[s.ServerID]", ServerClusters[PeerIdMap[peerId]], ServerClusters[s.ServerID])
		if peerId == PeerAddresses[s.ServerID] {
			promiseCount++
			log.Printf(prefix+"on %s Promise for self debug 1", s.ServerID)
			s.AppendToPrePrepareLogs(prePrepMsgData, seqNo)
			s.TxnState[clientRequest.Transaction.Index] = PP_STATE
			s.TxnStateBySeq[seqNo] = PP_STATE

			sig, _ := SignData(prePrepMsgData.ClientRequestHash, s.PrivateKey)
			prepMsgData := s.CreatePrepareMessage(prePrepMsgData, sig)
			signatureArr = append(signatureArr, prepMsgData.SignedMsg)
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

				resp, err := peer.PBFTPrePrepare(ctx, prePrepMsgData)
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
			isValid, _ := VerifyData(prePrepMsgData.ClientRequestHash, resp.SignedMsg.Signature, s.PubKeys[resp.SignedMsg.NodeId])
			if isValid {
				signatureArr = append(signatureArr, resp.SignedMsg)
				promiseCount++
			}
		}
	}

	// s.AppendToPrepareLogs(prepMsgData, seqNo)

	if len(errorCh) > 0 {
		return nil, promiseCount, fmt.Errorf("PrePrepare phase failed due to errors in peer requests")
	}

	log.Printf(prefix+"Promise count for %s : %d / %d -> %d", s.ServerID, promiseCount, s.ReplicaCount, 2*F+1)

	if promiseCount < 2*F+1 {
		return nil, promiseCount, fmt.Errorf("PrePrepare phase failed")
	}
	prepMsgData := &pb.PrepareRequest{}
	if !ByzSvrMap[s.ServerID] {
		prepMsgData = s.CreatePrepareRequestCertificate(prePrepMsgData, signatureArr)
	} else {
		log.Printf(prefix+"Prepare rejected for self and others as malicious node is leader(byz case) on %s", s.ServerID)
	}
	return prepMsgData, promiseCount, nil
}

func (s *TPCServer) PBFTPrePrepare(ctx context.Context, msg *pb.PrePrepareMessage) (*pb.PrepareMessage, error) {
	log.Printf("PrePrepareRequest :: Replica %s received client request, starting view change timer for txn %d", s.ServerID, msg.Request.Transaction.Index)
	// s.startViewChangeTimer()

	primary := ClusterMapping[ServerClusters[s.ServerID]].Leader //AllServers[msg.ViewNumber%s.ReplicaCount]
	log.Printf("----------->PrePrepareRequest from %s on %s", primary, s.ServerID)
	isValid, err := s.ValidatePrePrepareMsg(primary, msg)
	if err != nil {
		log.Printf("ValidatePrePrepareMsg  Error %v", err)
	}
	if !isValid {
		return nil, fmt.Errorf("Invalid PrePrepare message")
	}

	s.Mu.Lock()
	s.AppendToPrePrepareLogs(msg, msg.SequenceNumber)
	s.Mu.Unlock()

	if ByzSvrMap[s.ServerID] {
		return nil, fmt.Errorf("PrePrepare rejected by replica (Byzantine)")
	}

	// digest, _ := HashStruct(msg.Request)
	sig, _ := SignData(msg.ClientRequestHash, s.PrivateKey)
	return s.CreatePrepareMessage(msg, sig), nil
}
func (s *TPCServer) CreatePrepareMessage(msg *pb.PrePrepareMessage, sign string) *pb.PrepareMessage {
	signedmsg := &pb.SignedMessage{
		NodeId:    s.ServerID,
		Signature: sign,
	}
	// digest, _ := HashStruct(msg.Request)
	prepareMessage := &pb.PrepareMessage{
		ViewNumber:        s.ViewNumber,
		SequenceNumber:    msg.SequenceNumber,
		ClientRequestHash: msg.ClientRequestHash,
		SignedMsg:         signedmsg,
	}

	return prepareMessage
}

func (s *TPCServer) CreatePrepareRequestCertificate(msg *pb.PrePrepareMessage, signatureArr []*pb.SignedMessage) *pb.PrepareRequest {
	// digest, _ := HashStruct(msg.Request)
	signedSigMap := map[string]string{}

	for i, _ := range signatureArr {
		signedSigMap[signatureArr[i].NodeId] = signatureArr[i].Signature
	}
	prepareMessage := &pb.PrepareRequest{
		ViewNumber:     msg.ViewNumber,
		SequenceNumber: msg.SequenceNumber,
		Digest:         msg.ClientRequestHash,
		Signatures:     signedSigMap,
	}
	log.Printf("CreatePrepareRequestCertificate to sign :: %v ", prepareMessage)
	sign, _ := SignData(prepareMessage, s.PrivateKey)
	prepareMessage.LeaderSignature = sign

	return prepareMessage
}
