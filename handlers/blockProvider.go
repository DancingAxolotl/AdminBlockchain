package handlers

import (
	"log"
	"net/rpc"
)

// IBlockProvider interface for block providers
type IBlockProvider interface {
	GetBlockHeight() int
	GetBlock(index int) SignedBlockData
}

// RPCBlockProvider fetches blocks by rpc
type RPCBlockProvider struct {
	Client *rpc.Client
}

// GetBlock gets block with specified ID
func (bp RPCBlockProvider) GetBlock(index int) SignedBlockData {
	var block SignedBlockData
	err := bp.Client.Call("BlockPropagationHandler.GetBlock", index, &block)
	if err != nil {
		log.Print(err)
	}
	return block
}

// GetBlockHeight gets block with specified ID
func (bp RPCBlockProvider) GetBlockHeight() int {
	var block int
	err := bp.Client.Call("BlockPropagationHandler.GetBlockHeight", 0, &block)
	if err != nil {
		log.Print(err)
	}
	return block
}
