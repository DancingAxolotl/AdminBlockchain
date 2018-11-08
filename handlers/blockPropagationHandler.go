package handlers

import (
	"AdminBlockchain/storage"
	"errors"
	"log"
)

// BlockPropagationHandler for syncing clients with the blockchain.
type BlockPropagationHandler struct {
	Storage *storage.Provider
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
func (bp *BlockPropagationHandler) GetBlock(index int, block *storage.Block) error {
	var err = bp.checkState()
	if err == nil && index >= len(bp.Storage.Chain) {
		err = errors.New("Invalid index")
	}

	if err == nil {
		(*block).ID = bp.Storage.Chain[index].ID
		(*block).Data = bp.Storage.Chain[index].Data
		(*block).PrevHash = make([]byte, len(bp.Storage.Chain[index].PrevHash))
		for i, item := range bp.Storage.Chain[index].PrevHash {
			(*block).PrevHash[i] = item
		}
	}
	return err
}
