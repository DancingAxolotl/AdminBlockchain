package network

import (
	"AdminBlockchain/utils"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

// ServerNetworkProvider provides network access for server
type ServerNetworkProvider struct {
	listener   net.Listener
	rpcServer  *rpc.Server
	httpServer *http.Server
	running    bool
}

// NewServerProvider Create a new network provider
func NewServerProvider() ServerNetworkProvider {
	var np ServerNetworkProvider
	np.rpcServer = rpc.NewServer()
	np.httpServer = &http.Server{}
	np.running = false
	return np
}

// RegisterHandler registers handlers for rpc
func (np *ServerNetworkProvider) RegisterHandler(handler interface{}) {
	np.rpcServer.Register(handler)
}

// Start starts the server
func (np *ServerNetworkProvider) Start(addr string, port string) {
	np.rpcServer.HandleHTTP("/_goRPC_", "/debug/rpc")
	var err error
	np.listener, err = net.Listen("tcp", addr+":"+port)
	utils.LogErrorF(err)
	log.Printf("Serving RPC server on %v", addr+":"+port)

	// Start accept incoming HTTP connections
	err = np.httpServer.Serve(np.listener)
	utils.LogErrorF(err)

	np.running = true
}

// Stop stops the server
func (np *ServerNetworkProvider) Stop() {
	if np.running {
		err := np.listener.Close()
		utils.LogErrorF(err)

		np.rpcServer.Accept(np.listener)

		err = np.httpServer.Close()
		utils.LogErrorF(err)

		np.running = false
	}
}
