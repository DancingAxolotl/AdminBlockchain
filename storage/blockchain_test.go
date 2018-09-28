package blockchain

import "testing"

func assertEq(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Errorf("%s != %s", expected, actual)
	}
}

func assertArrayEq(t *testing.T, expected []byte, actual []byte) {
    for i, item := range expected {
        if item != actual[i] {
            t.Errorf("%s != %s", expected, actual)
        }
    }
}

// Check if a block is added to the blockchain
func TestAddBlock(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	lastBlock := blockchain[len(blockchain)-1]
    
	assertEq(t, lastBlock.data, "hello")
}

// Check if the previous hash of first block is 0
func TestAddFirstBlock(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	lastBlock := blockchain[len(blockchain)-1]
    
	assertArrayEq(t, lastBlock.prevHash, []byte{0})
}

// Check if the previous hash is equal to the hash of previous block
func TestAddBlockHash(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	firstBlock := blockchain[len(blockchain)-1]
    
	blockchain.AddBlock("data")
	lastBlock := blockchain[len(blockchain)-1]
    
	assertArrayEq(t, lastBlock.prevHash, firstBlock.Hash())
}

// Check if the previous hash is equal to the hash of previous block
func TestValidBlockchain(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	blockchain.AddBlock("data")
    
	assertEq(t, blockchain.IsValid(), true)
}

// Check if the previous hash is equal to the hash of previous block
func TestInvalidBlockchain(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	blockchain.AddBlock("data")
    
	blockchain[len(blockchain)-2].data = "fake"
    
	assertEq(t, blockchain.IsValid(), false)
}