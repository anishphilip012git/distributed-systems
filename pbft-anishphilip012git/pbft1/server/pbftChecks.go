package server

import (
	"bytes"
	"log"
	pb "pbft1/proto"
	. "pbft1/security"

	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"
)

// VerifyAggPrepareMessage verifies the validity of the CommitMessage
func (s *PBFTSvr) VerifyAggPrepareMessage(aggPrepMessage *pb.PrepareReqMessage) bool {
	// signature validation
	// log.Printf("to verify :: %v", commitMessage)
	sign := aggPrepMessage.LeaderSignature
	req := aggPrepMessage.Request
	aggPrepMessage.Request = nil
	aggPrepMessage.LeaderSignature = ""
	primary := AllServers[s.ViewNumber%ReplicaCount]
	isValid, err := VerifyData(aggPrepMessage, sign, s.PubKeys[primary])
	if !isValid || err != nil {
		log.Printf("VerifyCommitMessage rejected due to incorrect message %v", err)
		return false
	}
	aggPrepMessage.LeaderSignature = sign
	aggPrepMessage.Request = req
	// 1. Check if the view number matches the current view number
	if aggPrepMessage.ViewNumber != s.ViewNumber {
		log.Printf("Invalid view number: got %d, expected %d", aggPrepMessage.ViewNumber, s.ViewNumber)
		return false
	}
	seqNo := aggPrepMessage.PrepareCertificate.SequenceNumber
	if seqNo != aggPrepMessage.SequenceNumber {
		log.Printf("Invalid sequenceNumber number: got %d, expected %d", seqNo, aggPrepMessage.SequenceNumber)
		return false
	}

	// 2. Check the validity of the PrepareCertificate
	if !s.VerifyPrepareCertificate(aggPrepMessage.PrepareCertificate) {
		log.Println("Invalid PrepareCertificate in commit message")
		return false
	}

	// 3. Verify the aggregated signature
	valid := s.VerifyAggregatedSignature(aggPrepMessage.PrepareCertificate, aggPrepMessage.AggregatedSignature)
	if !valid {
		log.Println("Invalid aggregated signature")
		return false
	}
	hashFromPrepare := aggPrepMessage.PrepareCertificate.PrepareMessages[0].ClientRequestHash
	log.Printf("VerifyCommitMessage  request %v", aggPrepMessage.Request)
	digest, _ := HashStruct(aggPrepMessage.Request)
	// log.Printf("digest 2 :: %s -> %v", digest, aggPrepMessage.Request)
	if hashFromPrepare != digest {
		log.Println("Invalid Client request message  in commit message: hashes not matching", digest, hashFromPrepare)
		return false
	}
	// 4. Check for duplicates (if applicable)
	if s.IsCommitMessageProcessed(aggPrepMessage) {
		log.Println("Duplicate commit message received")
		return false
	}

	// All checks passed
	return true
}

// VerifyPrepareCertificate checks if the PrepareCertificate is valid
func (s *PBFTSvr) VerifyPrepareCertificate(prepareCert *pb.PrepareCertificate) bool {

	// Check if the prepare messages in the certificate are valid
	if len(prepareCert.PrepareMessages) <= 2*F {
		log.Printf("PrepareCertificate does not meet quorum requirements: %d", len(prepareCert.PrepareMessages))
		return false
	}
	prepSeqNo, prepViewNo := prepareCert.PrepareMessages[0].SequenceNumber, prepareCert.PrepareMessages[0].ViewNumber
	for _, prep := range prepareCert.PrepareMessages {
		if prepSeqNo != prep.SequenceNumber || prepViewNo != prep.ViewNumber {
			return false
		}
	}

	// Additional checks can be implemented here as necessary
	return true
}

// VerifyAggregatedSignature verifies the aggregated signature against the PrepareCertificate
func (s *PBFTSvr) VerifyAggregatedSignature(prepareCert *pb.PrepareCertificate, aggregatedSig []byte) bool {
	// Reconstruct the message that was signed (e.g., a hash of the transactions)
	message := s.ReconstructAggSignFromPrepareCertificate(prepareCert)

	// Use the public keys and aggregated signature to verify
	valid := bytes.Equal(message, aggregatedSig)
	return valid
}

// VerifyAggregatedSignature verifies the aggregated signature against the PrepareCertificate

// IsCommitMessageProcessed checks if a commit message has already been processed
func (s *PBFTSvr) IsCommitMessageProcessed(commitMessage *pb.PrepareReqMessage) bool {
	// Implement logic to track processed commit messages, e.g., using a map
	// Here you would check if the commitMessage has already been handled.
	// This could be implemented with a map of processed commit message IDs.
	seqNo := commitMessage.PrepareCertificate.PrepareMessages[0].SequenceNumber
	_, ok := s.CommitedLogs[seqNo]
	return ok // Placeholder: Implement your logic to check for duplicates
}

// ReconstructMessageFromPrepareCertificate can reconstruct the message based on the PrepareCertificate
func (s *PBFTSvr) ReconstructAggSignFromPrepareCertificate(prepareCert *pb.PrepareCertificate) []byte {
	// You need to define how you reconstruct the message that was originally signed
	// For example, you could concatenate the view number, sequence number, and transaction data
	// For simplicity, let's say we're hashing the prepare messages
	// This is a placeholder and should be implemented according to your protocol logic
	partialSigs := make([]*ted25519.PartialSignature, 0, ReplicaCount)
	for _, prep := range prepareCert.PrepareMessages {
		// MsgWiseNonceQ[prep.SequenceNumber]
		ps := ted25519.PartialSignature{
			ShareIdentifier: prep.ReplicaPartialSignature[0],  // first byte is ShareIdentifier
			Sig:             prep.ReplicaPartialSignature[1:], // remaining bytes are Sig
		}
		partialSigs = append(partialSigs, &ps)
	}

	aggSign := s.AggregateSign(partialSigs)
	return aggSign
}
