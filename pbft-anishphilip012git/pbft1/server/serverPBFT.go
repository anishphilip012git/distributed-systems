package server

import (
	"context"
	"fmt"
	"log"
	pb "pbft1/proto"
	. "pbft1/security"
	"time"

	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"
)

var MsgWiseNonceQ map[int64]SharedNonce
var ChkpntMsgNonceQ map[int64]SharedNonce

// Define the timer duration for a view change trigger
const viewChangeTimeout = 10 * time.Second

const rpcCtxDeadlineTimeoutViewChangeCalls = 1 * time.Second
const rpcCtxDeadlineTimeout = 2 * time.Second
const rpcCtxDeadlineTimeoutForClient = 5 * time.Second

func (s *PBFTSvr) IniitiatePBFT(ctx context.Context, clientRequest *pb.ClientRequestMessage) (*pb.ClientReplyMessage, error) {
	// s.mu.Lock()
	// defer s.mu.Unlock()
	log.SetPrefix("IniitiatePBFT ::")
	idx := clientRequest.Transaction.Index
	msg := ""
	s.Mu.Lock()
	s.SequenceNumber++
	s.TxnState[idx] = X_STATE
	seqNo := s.SequenceNumber
	TxnIdMapWithSeqNo[idx] = seqNo
	s.Mu.Unlock()
	// Increment commit count for the sequence number
	log.Printf("for client request  at %s , %v ", s.ServerID, clientRequest)
	aggrgateSign, _, _ := s.PrePreparePhaseCall(clientRequest, seqNo)
	if len(aggrgateSign) == 0 {
		msg = "Prepare rejected"
		// time.Sleep(viewChangeTimeout)
	} else {
		cnt, _ := s.PreparePhaseCall(clientRequest, aggrgateSign, seqNo)
		if cnt < F+1 {
			msg = "Prepare rejected"
		} else {
			s.CommitPhaseCall(clientRequest, seqNo)
			msg = "Success"
		}
	}

	log.SetPrefix("")
	replyMessage := s.CreateClientReplyMessage(msg, s.ViewNumber, clientRequest.Transaction.Index)
	return replyMessage, nil

}

// func (s *PBFTSvr) CreateReplyMessage(msg string, viewNumber, idx int64) *pb.ReplyMessage {
// 	// digest, _ := HashStruct(msg.Request)
// 	reply := &pb.ReplyMessage{
// 		ViewNumber:       viewNumber,
// 		Status:           msg,
// 		RequestId:        idx,
// 		ReplicaSignature: "",
// 	}
// 	sign, _ := SignData(reply, s.PrivateKey)
// 	reply.ReplicaSignature = sign
// 	return reply
// }

func (s *PBFTSvr) AggregateSign(partialSigs []*ted25519.PartialSignature) []byte {
	// Step 5: Primary aggregates the collected partial signatures
	sig, err := AggregateSignatures(s.TSconfig.Config, partialSigs)
	if err != nil {
		fmt.Printf("Error aggregating signatures: %v\n", err)
		return nil
	}
	return sig

}

func (s *PBFTSvr) AggregateAndVerify(partialSigs []*ted25519.PartialSignature, message []byte) (bool, []byte) {
	log.Println("AggregateAndVerify :: no of ps ", len(partialSigs), string(message))
	// Step 5: Primary aggregates the collected partial signatures
	sig, err := AggregateSignatures(s.TSconfig.Config, partialSigs)
	if err != nil {
		fmt.Printf("Error aggregating signatures: %v\n", err)
		return false, nil
	}

	// Step 6: Verify the final aggregated signature
	valid, err := VerifySignature(s.TSconfig.Pub, message, sig)
	if err != nil {
		fmt.Printf("Error in signature verification: %v\n", err)
	}
	return valid, sig
}
