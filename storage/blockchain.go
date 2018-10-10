package storage

import (
    "bytes"
    "crypto/sha256"
)

// Block is a basic block within a blockchain
type Block struct {
    id       int
    prevHash []byte
    data     string
}

// Hash function, computes the hash of the block
func (block Block) Hash() []byte {
    hash := sha256.New()

    var buffer bytes.Buffer
    buffer.WriteString(block.data)
    buffer.WriteString(string(block.id))
    buffer.Write(block.prevHash)

    hash.Write(buffer.Bytes())
    return hash.Sum(nil)
}

// Blockchain is a chain of blocks
type Blockchain []Block


// AddBlock adds a block to the blockchain
func (blockchain *Blockchain) AddBlock(data string) {
    hash, blockHeight := []byte{0}, len(*blockchain)
    
    if blockHeight > 0 {
        hash = (*blockchain)[blockHeight - 1].Hash()
    }
    
    *blockchain = append(*blockchain, Block{len(*blockchain), hash, data})
}

// IsValid Checks if the blockchain is valid.
func (blockchain *Blockchain) IsValid() bool {
    blockHeight := len(*blockchain)
    if blockHeight == 0 {
        return true
    }
    lastBlock := (*blockchain)[0]
    
    // check if first block is valid
    if len(lastBlock.prevHash) != 1 || lastBlock.prevHash[0] != 0 {
        return false
    }
    
    for i := 1; i < len(*blockchain); i++ {
        nextBlock := (*blockchain)[i]
        hash := lastBlock.Hash()
        for j, item := range nextBlock.prevHash {
            if item != hash[j] {
                return false
            }
        }
        
        lastBlock = nextBlock
    }
    
    return true
} 
