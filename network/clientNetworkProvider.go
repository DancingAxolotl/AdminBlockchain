package network

import (
	"AdminBlockchain/utils"
	"net/rpc"
)

// ClientNetworkProvider provides network access for client
type ClientNetworkProvider struct {
	client *rpc.Client
}

// Start starts the server
func (np *ClientNetworkProvider) Start(addr string, port string) {
	var err error
	np.client, err = rpc.DialHTTP("tcp", addr+":"+port)
	utils.LogErrorF(err)
}

// Stop stops the server
func (np *ClientNetworkProvider) Stop() {
	np.client.Close()
}
