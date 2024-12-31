package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ukibbb/blockchain-go/types"
)

func TestNewBlockChain(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(1))
	assert.False(t, bc.HasBlock(100))

}

func TestAddBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlockWithSignature(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}
	assert.Equal(t, bc.Height(), uint32(lenBlocks))

	assert.NotNil(t, bc.AddBlock(RandomBlockWithSignature(t, 89, types.Hash{})))
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	assert.Nil(t, bc.AddBlock(RandomBlockWithSignature(t, 1, getPrevBlockHash(t, bc, uint32(1)))))
	assert.NotNil(t, bc.AddBlock(RandomBlockWithSignature(t, 3, types.Hash{})))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlockWithSignature(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}

}

func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(RandomBlockWithSignature(t, 0, types.Hash{}))

	assert.Nil(t, err)
	return bc
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)

}