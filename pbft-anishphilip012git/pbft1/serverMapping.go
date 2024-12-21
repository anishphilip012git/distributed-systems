package main

import (
	"sync"
	// Import the generated proto package
)

type ServerMappingAtClient struct {
	name string
	addr string
	// balance  map[string]int
	// localLog []Transaction
	mu sync.Mutex
	// isDown   bool // Server down status
	// buffer   []Transaction
}

// NewServerMapping creates a new server with an initial balance
func NewServerMapping(name string, addr string) *ServerMappingAtClient {
	svr := &ServerMappingAtClient{
		name: name,
		addr: addr,
	}
	return svr
}
