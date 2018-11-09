package handlers

import (
	"AdminBlockchain/storage"
	"AdminBlockchain/utils"
	"errors"
	"log"
)

// BlockPropagationHandler for syncing clients with the blockchain.
type BlockPropagationHandler struct {
	Signer  utils.SignatureCreator
	Storage *storage.Provider
}

// SignedBlockData block data signed with private key of the server
type SignedBlockData struct {
	BlockData storage.Block
	Signature []byte
}

func (bp *BlockPropagationHandler) checkState() error {
	if bp.Storage == nil {
		log.Print("Block propagation handler is not initialized")
		return errors.New("Handler is not initialized")
	}
	return nil
}

// GetBlockHeight rpc method, returns the current block height
func (bp *BlockPropagationHandler) GetBlockHeight(_, height *int) error {
	var err = bp.checkState()
	if err == nil {
		*height = len(bp.Storage.Chain)
	}
	return err
}

// GetBlock rpc method, returns specified block
func (bp *BlockPropagationHandler) GetBlock(index int, block *SignedBlockData) error {
	var err = bp.checkState()
	if err == nil && index > len(bp.Storage.Chain) {
		err = errors.New("Invalid index")
	}

	if err == nil {
		(*block).BlockData.ID = bp.Storage.Chain[index].ID
		(*block).BlockData.Data = bp.Storage.Chain[index].Data
		(*block).BlockData.PrevHash = make([]byte, len(bp.Storage.Chain[index].PrevHash))
		for i, item := range bp.Storage.Chain[index].PrevHash {
			(*block).BlockData.PrevHash[i] = item
		}
		(*block).Signature, err = bp.Signer.Sign((*block).BlockData.Hash())
		utils.LogError(err)
	}
	return err
}
