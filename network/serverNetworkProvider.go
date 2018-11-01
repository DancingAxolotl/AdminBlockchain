package network

import (
	"AdminBlockchain/handlers"
	"net"
	"net/rpc"
)

// ServerNetworkProvider provides network access for server
type ServerNetworkProvider struct {
	listener net.Listener
	server   *rpc.Server
}

// NewServerProvider Create a new network provider
func NewServerProvider() ServerNetworkProvider {
	var np ServerNetworkProvider
	np.server = rpc.NewServer()
	return np
}

// RegisterHandler registers handlers for rpc
func (np *ServerNetworkProvider) RegisterHandler(handler handlers.IHandler) {
	np.server.Register(handler)
}

// Start starts the server
func (np *ServerNetworkProvider) Start(addr string, port string) {
	np.listener, _ = net.Listen("tcp", addr+":"+port)
	np.server.Accept(np.listener)
}

// Stop stops the server
func (np *ServerNetworkProvider) Stop() {
	np.listener.Close()
}
