package handlers

import (
	"AdminBlockchain/storage"
	"errors"
)

// BlockSyncHandler handles fetching blocks
type BlockSyncHandler struct {
	QueryHandler *SimpleQueryHandler
}

func (sync *BlockSyncHandler) PerformSyncRpc(rpc RPCBlockProvider) error {
	
}

// Sync loads new blocks from a blockProvider
func (sync *BlockSyncHandler) Sync(blockProvider IBlockProvider) error {
	if sync.QueryHandler == nil {
		return errors.New("handler not initialized")
	}

	localHeight := len(sync.QueryHandler.Sp.Chain)
	externalHeight := blockProvider.GetBlockHeight()
	for localHeight != externalHeight {
		block := blockProvider.GetBlock(localHeight)
		err := sync.pushBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

// pushBlock validates and adds block to the blockchain
func (sync *BlockSyncHandler) pushBlock(block storage.Block) error {
	if sync.QueryHandler == nil {
		return errors.New("handler not initialized")
	}

	if block.ID != len(sync.QueryHandler.Sp.Chain) {
		return errors.New("invalid block id, push only at block height")
	}

	validationChain := append(sync.QueryHandler.Sp.Chain, block)
	if validationChain.IsValid() {
		sync.QueryHandler.Sp.Chain = validationChain
		sync.QueryHandler.AcceptBlock(block)
	}
	return nil
}
