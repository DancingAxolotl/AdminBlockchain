package storage

import (
	"bytes"
	"crypto/sha256"
)

// Block is a basic block within a blockchain
type Block struct {
	Id       int
	PrevHash []byte
	Data     string
}

// Hash function, computes the hash of the block
func (block Block) Hash() []byte {
	hash := sha256.New()

	var buffer bytes.Buffer
	buffer.WriteString(block.Data)
	buffer.WriteString(string(block.Id))
	buffer.Write(block.PrevHash)

	hash.Write(buffer.Bytes())
	return hash.Sum(nil)
}

// Blockchain is a chain of blocks
type Blockchain []Block

// AddBlock adds a block to the blockchain
func (blockchain *Blockchain) AddBlock(data string) {
	hash, blockHeight := []byte{0}, len(*blockchain)

	if blockHeight > 0 {
		hash = (*blockchain)[blockHeight-1].Hash()
	}

	*blockchain = append(*blockchain, Block{blockHeight + 1, hash, data})
}

//InsertBlock attempts to insert a block at the end of the blokchain. It doesn't check if the hash of the previous block is valid.
// After inserting any amount of blocks make sure to check the validity of the chain.
func (blockchain *Blockchain) InsertBlock(block Block) {
	*blockchain = append(*blockchain, block)
}

// IsValid Checks if the blockchain is valid.
func (blockchain *Blockchain) IsValid() bool {
	blockHeight := len(*blockchain)
	if blockHeight == 0 {
		return true
	}
	lastBlock := (*blockchain)[0]

	// check if first block is valid
	if len(lastBlock.PrevHash) != 1 || lastBlock.PrevHash[0] != 0 {
		return false
	}

	for i := 1; i < blockHeight; i++ {
		nextBlock := (*blockchain)[i]
		hash := lastBlock.Hash()
		for j, item := range nextBlock.PrevHash {
			if item != hash[j] {
				return false
			}
		}

		lastBlock = nextBlock
	}

	return true
}
