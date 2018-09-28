package blockchain

import "testing"

func assertEq(t *testing.T, expected interface, actual interface, description string = "")
{
	if expected != actual {
		t.Error(
			"For", description,
			"expected", expected,
			"got", actual)
	}
}

func TestAddBlock(t *testing.T) {
	blockchain := Blockchain{}
	blockchain.AddBlock("hello")
	lastBlock := blockchain[len(blockchain)-1]
    
	assertEq(t, lastBlock.data, "hello", "last block data")
}