package handlers

import (
	"AdminBlockchain/storage"
	"log"
	"net/rpc"
)

// IBlockProvider interface for block providers
type IBlockProvider interface {
	GetBlockHeight() int
	GetBlock(index int) storage.Block
}

// RpcBlockProvider fetches blocks by rpc
type RPCBlockProvider struct {
	client *rpc.Client
}

// GetBlock gets block with specified ID
func (bp RPCBlockProvider) GetBlock(index int) storage.Block {
	var block storage.Block
	err := bp.client.Call("BlockPropagationHandler.GetBlock", index, &block)
	if err != nil {
		log.Print(err)
	}
	return block
}

// GetBlockHeight gets block with specified ID
func (bp RPCBlockProvider) GetBlockHeight() int {
	var block int
	err := bp.client.Call("BlockPropagationHandler.GetBlockHeight", 0, &block)
	if err != nil {
		log.Print(err)
	}
	return block
}
