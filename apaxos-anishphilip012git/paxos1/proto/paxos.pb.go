// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.2
// source: proto/paxos.proto

package paxos

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_proto_paxos_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{0}
}

type HealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Healthy bool `protobuf:"varint,1,opt,name=healthy,proto3" json:"healthy,omitempty"`
}

func (x *HealthResponse) Reset() {
	*x = HealthResponse{}
	mi := &file_proto_paxos_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthResponse) ProtoMessage() {}

func (x *HealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthResponse.ProtoReflect.Descriptor instead.
func (*HealthResponse) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{1}
}

func (x *HealthResponse) GetHealthy() bool {
	if x != nil {
		return x.Healthy
	}
	return false
}

// A message to represent a transaction in the format <A, B, amt>
type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sender   string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`     // A
	Receiver string `protobuf:"bytes,2,opt,name=receiver,proto3" json:"receiver,omitempty"` // B
	Amount   int64  `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`    // amt
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	mi := &file_proto_paxos_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{2}
}

func (x *Transaction) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *Transaction) GetReceiver() string {
	if x != nil {
		return x.Receiver
	}
	return ""
}

func (x *Transaction) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

// TransactionRequest is the request message for sending a transaction
type TransactionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionEntry map[int64]*Transaction `protobuf:"bytes,1,rep,name=transaction_entry,json=transactionEntry,proto3" json:"transaction_entry,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // The transaction to be sent
}

func (x *TransactionRequest) Reset() {
	*x = TransactionRequest{}
	mi := &file_proto_paxos_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransactionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionRequest) ProtoMessage() {}

func (x *TransactionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionRequest.ProtoReflect.Descriptor instead.
func (*TransactionRequest) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{3}
}

func (x *TransactionRequest) GetTransactionEntry() map[int64]*Transaction {
	if x != nil {
		return x.TransactionEntry
	}
	return nil
}

// TransactionResponse is the response message for transaction processing
type TransactionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"` // Indicates if the transaction was processed successfully
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`  // Optional message for additional information
}

func (x *TransactionResponse) Reset() {
	*x = TransactionResponse{}
	mi := &file_proto_paxos_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransactionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionResponse) ProtoMessage() {}

func (x *TransactionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionResponse.ProtoReflect.Descriptor instead.
func (*TransactionResponse) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{4}
}

func (x *TransactionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *TransactionResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type PrepareRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BallotNumber        int64  `protobuf:"varint,1,opt,name=ballot_number,json=ballotNumber,proto3" json:"ballot_number,omitempty"`
	LeaderId            string `protobuf:"bytes,2,opt,name=leader_id,json=leaderId,proto3" json:"leader_id,omitempty"`
	LastCommittedBallot int64  `protobuf:"varint,3,opt,name=last_committed_ballot,json=lastCommittedBallot,proto3" json:"last_committed_ballot,omitempty"` // int64 last_svr_balance=4;
}

func (x *PrepareRequest) Reset() {
	*x = PrepareRequest{}
	mi := &file_proto_paxos_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrepareRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrepareRequest) ProtoMessage() {}

func (x *PrepareRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrepareRequest.ProtoReflect.Descriptor instead.
func (*PrepareRequest) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{5}
}

func (x *PrepareRequest) GetBallotNumber() int64 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

func (x *PrepareRequest) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

func (x *PrepareRequest) GetLastCommittedBallot() int64 {
	if x != nil {
		return x.LastCommittedBallot
	}
	return 0
}

type PromiseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Promise                 bool                   `protobuf:"varint,1,opt,name=promise,proto3" json:"promise,omitempty"`
	LastAcceptedBallot      int64                  `protobuf:"varint,2,opt,name=last_accepted_ballot,json=lastAcceptedBallot,proto3" json:"last_accepted_ballot,omitempty"`
	LastAcceptedValue       map[int64]*Transaction `protobuf:"bytes,3,rep,name=last_accepted_value,json=lastAcceptedValue,proto3" json:"last_accepted_value,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	UncommittedTransactions map[int64]*Transaction `protobuf:"bytes,4,rep,name=uncommitted_transactions,json=uncommittedTransactions,proto3" json:"uncommitted_transactions,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Updated to use the Transaction format
}

func (x *PromiseResponse) Reset() {
	*x = PromiseResponse{}
	mi := &file_proto_paxos_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PromiseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PromiseResponse) ProtoMessage() {}

func (x *PromiseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PromiseResponse.ProtoReflect.Descriptor instead.
func (*PromiseResponse) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{6}
}

func (x *PromiseResponse) GetPromise() bool {
	if x != nil {
		return x.Promise
	}
	return false
}

func (x *PromiseResponse) GetLastAcceptedBallot() int64 {
	if x != nil {
		return x.LastAcceptedBallot
	}
	return 0
}

func (x *PromiseResponse) GetLastAcceptedValue() map[int64]*Transaction {
	if x != nil {
		return x.LastAcceptedValue
	}
	return nil
}

func (x *PromiseResponse) GetUncommittedTransactions() map[int64]*Transaction {
	if x != nil {
		return x.UncommittedTransactions
	}
	return nil
}

type AcceptRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BallotNumber         int64                  `protobuf:"varint,1,opt,name=ballot_number,json=ballotNumber,proto3" json:"ballot_number,omitempty"`
	AccpetedTransactions map[int64]*Transaction `protobuf:"bytes,2,rep,name=accpeted_transactions,json=accpetedTransactions,proto3" json:"accpeted_transactions,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Updated to handle a list of transactions
}

func (x *AcceptRequest) Reset() {
	*x = AcceptRequest{}
	mi := &file_proto_paxos_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AcceptRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcceptRequest) ProtoMessage() {}

func (x *AcceptRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcceptRequest.ProtoReflect.Descriptor instead.
func (*AcceptRequest) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{7}
}

func (x *AcceptRequest) GetBallotNumber() int64 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

func (x *AcceptRequest) GetAccpetedTransactions() map[int64]*Transaction {
	if x != nil {
		return x.AccpetedTransactions
	}
	return nil
}

type AcceptedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Accepted bool `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
}

func (x *AcceptedResponse) Reset() {
	*x = AcceptedResponse{}
	mi := &file_proto_paxos_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AcceptedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcceptedResponse) ProtoMessage() {}

func (x *AcceptedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcceptedResponse.ProtoReflect.Descriptor instead.
func (*AcceptedResponse) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{8}
}

func (x *AcceptedResponse) GetAccepted() bool {
	if x != nil {
		return x.Accepted
	}
	return false
}

type DecisionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BallotNumber        int64                  `protobuf:"varint,1,opt,name=ballot_number,json=ballotNumber,proto3" json:"ballot_number,omitempty"`
	TransactionsEntries map[int64]*Transaction `protobuf:"bytes,2,rep,name=transactions_entries,json=transactionsEntries,proto3" json:"transactions_entries,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Updated to reflect the decision on a list of transactions
}

func (x *DecisionRequest) Reset() {
	*x = DecisionRequest{}
	mi := &file_proto_paxos_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DecisionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecisionRequest) ProtoMessage() {}

func (x *DecisionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecisionRequest.ProtoReflect.Descriptor instead.
func (*DecisionRequest) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{9}
}

func (x *DecisionRequest) GetBallotNumber() int64 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

func (x *DecisionRequest) GetTransactionsEntries() map[int64]*Transaction {
	if x != nil {
		return x.TransactionsEntries
	}
	return nil
}

type DecisionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *DecisionResponse) Reset() {
	*x = DecisionResponse{}
	mi := &file_proto_paxos_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DecisionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecisionResponse) ProtoMessage() {}

func (x *DecisionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_paxos_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecisionResponse.ProtoReflect.Descriptor instead.
func (*DecisionResponse) Descriptor() ([]byte, []int) {
	return file_proto_paxos_proto_rawDescGZIP(), []int{10}
}

func (x *DecisionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_proto_paxos_proto protoreflect.FileDescriptor

var file_proto_paxos_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x2a, 0x0a, 0x0e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x22,
	0x59, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76,
	0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0xcb, 0x01, 0x0a, 0x12, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x5c, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x70,
	0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x10, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x1a,
	0x57, 0x0a, 0x15, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x78, 0x6f,
	0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x49, 0x0a, 0x13, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x0e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x62,
	0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1b, 0x0a, 0x09, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x15, 0x6c, 0x61, 0x73, 0x74,
	0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x6c, 0x6f,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x13, 0x6c, 0x61, 0x73, 0x74, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x22, 0xe6, 0x03, 0x0a,
	0x0f, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x61,
	0x73, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x6c,
	0x6f, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x12, 0x5d, 0x0a, 0x13,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x70, 0x61, 0x78, 0x6f,
	0x73, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x4c, 0x61, 0x73, 0x74, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x11, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x6e, 0x0a, 0x18, 0x75,
	0x6e, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e,
	0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x55, 0x6e, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65,
	0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x17, 0x75, 0x6e, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x58, 0x0a, 0x16, 0x4c,
	0x61, 0x73, 0x74, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x5e, 0x0a, 0x1c, 0x55, 0x6e, 0x63, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x74, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xf6, 0x01, 0x0a, 0x0d, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x62, 0x61, 0x6c, 0x6c, 0x6f,
	0x74, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c,
	0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x63, 0x0a, 0x15,
	0x61, 0x63, 0x63, 0x70, 0x65, 0x74, 0x65, 0x64, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x70, 0x61,
	0x78, 0x6f, 0x73, 0x2e, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x2e, 0x41, 0x63, 0x63, 0x70, 0x65, 0x74, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x14, 0x61, 0x63, 0x63,
	0x70, 0x65, 0x74, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x1a, 0x5b, 0x0a, 0x19, 0x41, 0x63, 0x63, 0x70, 0x65, 0x74, 0x65, 0x64, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x2e,
	0x0a, 0x10, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x22, 0xf6,
	0x01, 0x0a, 0x0f, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x5f, 0x6e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f,
	0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x62, 0x0a, 0x14, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x44, 0x65,
	0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x13, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x1a, 0x5a, 0x0a, 0x18, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73,
	0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x2c, 0x0a, 0x10, 0x44, 0x65, 0x63, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0xfb, 0x02, 0x0a, 0x05, 0x50, 0x61, 0x78, 0x6f, 0x73, 0x12,
	0x38, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x12, 0x15, 0x2e, 0x70, 0x61, 0x78,
	0x6f, 0x73, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x16, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x41, 0x63, 0x63,
	0x65, 0x70, 0x74, 0x12, 0x14, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x41, 0x63, 0x63, 0x65,
	0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x61, 0x78, 0x6f,
	0x73, 0x2e, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x39, 0x0a, 0x06, 0x44, 0x65, 0x63, 0x69, 0x64, 0x65, 0x12, 0x16, 0x2e, 0x70,
	0x61, 0x78, 0x6f, 0x73, 0x2e, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x44, 0x65, 0x63,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a,
	0x0f, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x19, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x70, 0x61,
	0x78, 0x6f, 0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x0c, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x1a, 0x15, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x48, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x46, 0x0a, 0x0d, 0x53,
	0x79, 0x6e, 0x63, 0x44, 0x61, 0x74, 0x61, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x19, 0x2e, 0x70,
	0x61, 0x78, 0x6f, 0x73, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x42, 0x17, 0x5a, 0x15, 0x2e, 0x2e, 0x2f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x31,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_paxos_proto_rawDescOnce sync.Once
	file_proto_paxos_proto_rawDescData = file_proto_paxos_proto_rawDesc
)

func file_proto_paxos_proto_rawDescGZIP() []byte {
	file_proto_paxos_proto_rawDescOnce.Do(func() {
		file_proto_paxos_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_paxos_proto_rawDescData)
	})
	return file_proto_paxos_proto_rawDescData
}

var file_proto_paxos_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_proto_paxos_proto_goTypes = []any{
	(*Empty)(nil),               // 0: paxos.Empty
	(*HealthResponse)(nil),      // 1: paxos.HealthResponse
	(*Transaction)(nil),         // 2: paxos.Transaction
	(*TransactionRequest)(nil),  // 3: paxos.TransactionRequest
	(*TransactionResponse)(nil), // 4: paxos.TransactionResponse
	(*PrepareRequest)(nil),      // 5: paxos.PrepareRequest
	(*PromiseResponse)(nil),     // 6: paxos.PromiseResponse
	(*AcceptRequest)(nil),       // 7: paxos.AcceptRequest
	(*AcceptedResponse)(nil),    // 8: paxos.AcceptedResponse
	(*DecisionRequest)(nil),     // 9: paxos.DecisionRequest
	(*DecisionResponse)(nil),    // 10: paxos.DecisionResponse
	nil,                         // 11: paxos.TransactionRequest.TransactionEntryEntry
	nil,                         // 12: paxos.PromiseResponse.LastAcceptedValueEntry
	nil,                         // 13: paxos.PromiseResponse.UncommittedTransactionsEntry
	nil,                         // 14: paxos.AcceptRequest.AccpetedTransactionsEntry
	nil,                         // 15: paxos.DecisionRequest.TransactionsEntriesEntry
}
var file_proto_paxos_proto_depIdxs = []int32{
	11, // 0: paxos.TransactionRequest.transaction_entry:type_name -> paxos.TransactionRequest.TransactionEntryEntry
	12, // 1: paxos.PromiseResponse.last_accepted_value:type_name -> paxos.PromiseResponse.LastAcceptedValueEntry
	13, // 2: paxos.PromiseResponse.uncommitted_transactions:type_name -> paxos.PromiseResponse.UncommittedTransactionsEntry
	14, // 3: paxos.AcceptRequest.accpeted_transactions:type_name -> paxos.AcceptRequest.AccpetedTransactionsEntry
	15, // 4: paxos.DecisionRequest.transactions_entries:type_name -> paxos.DecisionRequest.TransactionsEntriesEntry
	2,  // 5: paxos.TransactionRequest.TransactionEntryEntry.value:type_name -> paxos.Transaction
	2,  // 6: paxos.PromiseResponse.LastAcceptedValueEntry.value:type_name -> paxos.Transaction
	2,  // 7: paxos.PromiseResponse.UncommittedTransactionsEntry.value:type_name -> paxos.Transaction
	2,  // 8: paxos.AcceptRequest.AccpetedTransactionsEntry.value:type_name -> paxos.Transaction
	2,  // 9: paxos.DecisionRequest.TransactionsEntriesEntry.value:type_name -> paxos.Transaction
	5,  // 10: paxos.Paxos.Prepare:input_type -> paxos.PrepareRequest
	7,  // 11: paxos.Paxos.Accept:input_type -> paxos.AcceptRequest
	9,  // 12: paxos.Paxos.Decide:input_type -> paxos.DecisionRequest
	3,  // 13: paxos.Paxos.SendTransaction:input_type -> paxos.TransactionRequest
	0,  // 14: paxos.Paxos.CheckHealth:input_type -> paxos.Empty
	3,  // 15: paxos.Paxos.SyncDataStore:input_type -> paxos.TransactionRequest
	6,  // 16: paxos.Paxos.Prepare:output_type -> paxos.PromiseResponse
	8,  // 17: paxos.Paxos.Accept:output_type -> paxos.AcceptedResponse
	10, // 18: paxos.Paxos.Decide:output_type -> paxos.DecisionResponse
	4,  // 19: paxos.Paxos.SendTransaction:output_type -> paxos.TransactionResponse
	1,  // 20: paxos.Paxos.CheckHealth:output_type -> paxos.HealthResponse
	4,  // 21: paxos.Paxos.SyncDataStore:output_type -> paxos.TransactionResponse
	16, // [16:22] is the sub-list for method output_type
	10, // [10:16] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_proto_paxos_proto_init() }
func file_proto_paxos_proto_init() {
	if File_proto_paxos_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_paxos_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_paxos_proto_goTypes,
		DependencyIndexes: file_proto_paxos_proto_depIdxs,
		MessageInfos:      file_proto_paxos_proto_msgTypes,
	}.Build()
	File_proto_paxos_proto = out.File
	file_proto_paxos_proto_rawDesc = nil
	file_proto_paxos_proto_goTypes = nil
	file_proto_paxos_proto_depIdxs = nil
}