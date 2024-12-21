package server

import (
	"context"
	"fmt"
	"log"
	pb "tpcbyz/proto"
)

func (s *TPCServer) ClientRequestTPC(ctx context.Context, req *pb.ClientRequestMessage) (*pb.ClientReplyMessage, error) {
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
		log.Printf("Primary %s received client request, for txn %d", s.ServerID, req.Transaction.Index)

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
