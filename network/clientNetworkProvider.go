package network

import (
	"log"
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
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
}

// Stop stops the server
func (np *ClientNetworkProvider) Stop() {
	np.client.Close()
}
