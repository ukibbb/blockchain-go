package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type BlockChain struct {
	store   Storage
	headers []*Header

	validator Validator
}

func NewBlockChain(genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		store: NewMemoryStore(),
	}
	bc.SetValidator(NewBlockValidator(bc))
	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *BlockChain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *BlockChain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}

	return bc.addBlockWithoutValidation(b)
}

func (bc *BlockChain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("givent height (%d) to high", height)
	}
	return bc.headers[height], nil
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	logrus.WithFields(logrus.Fields{
		"height": b.Height,
		"hash":   b.Hash(BlockHasher{}),
	}).Info("adding new block")
	bc.headers = append(bc.headers, b.Header)
	return bc.store.Put(b)

}

func (bc *BlockChain) addGenesisBlock(genesis *Block) {

}
