package main

import (
	"context"
	"log"
	"time"

	pb "paxos1/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id           string
	server       *ServerMappingAtClient
	conn         pb.PaxosClient // gRPC client for communication
	bufferClient []Transaction  //if unable to buffer at server
}

// NewClient creates a new client and maps it to a specific server
func NewClient(id string, server *ServerMappingAtClient) *Client {
	conn, err := grpc.NewClient(server.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := pb.NewPaxosClient(conn)

	return &Client{
		id: id,
		// server: server,
		conn: c,
	}
}

// SendTransaction sends a transaction to the client's assigned server via gRPC
func (c *Client) SendTransaction(index int64, tx Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	txnEntry := map[int64]*pb.Transaction{
		index: {
			Sender:   tx.Sender,
			Receiver: tx.Receiver,
			Amount:   int64(tx.Amount),
		},
	}

	// Create the TransactionRequest object.
	request := &pb.TransactionRequest{
		TransactionEntry: txnEntry,
	}
	// Send transaction via gRPC
	resp, err := c.conn.SendTransaction(ctx, request)

	if err != nil {
		log.Printf("Client %s failed to send transaction to server %s: %v", c.id, c.server.name, err)
		// Retry or handle buffering logic here
		// c.server.BufferTransaction(tx)
		return
	}

	log.Printf("%s received response from server: %v\n", c.id, resp)
}
