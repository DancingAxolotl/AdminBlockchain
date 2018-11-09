package handlers

import (
	"AdminBlockchain/storage"
	"AdminBlockchain/utils"
	"errors"
)

// BlockSyncHandler handles fetching blocks
type BlockSyncHandler struct {
	StorageProvider *storage.Provider
	QueryHandlers   []IHandler
	SignValidator   utils.SignatureValidator
}

// Sync loads new blocks from a blockProvider
func (sync *BlockSyncHandler) Sync(blockProvider IBlockProvider) error {
	if sync.StorageProvider == nil {
		return errors.New("handler not initialized")
	}

	localHeight := len(sync.StorageProvider.Chain)
	externalHeight := blockProvider.GetBlockHeight()
	var err error
	for ; localHeight != externalHeight && err == nil; localHeight++ {
		block := blockProvider.GetBlock(localHeight)
		err = sync.SignValidator.CheckSignature(block.BlockData.Hash(), block.Signature)
		if err == nil {
			err = sync.pushBlock(block.BlockData)
		}
	}
	return err
}

// pushBlock validates and adds block to the blockchain
func (sync *BlockSyncHandler) pushBlock(block storage.Block) error {
	if sync.StorageProvider == nil {
		return errors.New("handler not initialized")
	}

	if block.ID != len(sync.StorageProvider.Chain) {
		return errors.New("invalid block id, push only at block height")
	}

	validationChain := append(sync.StorageProvider.Chain, block)
	if validationChain.IsValid() {
		sync.StorageProvider.Chain = validationChain
		for _, handler := range sync.QueryHandlers {
			if handler != nil {
				handler.AcceptBlock(block)
			}
		}
	}
	return nil
}
