package core

import "fmt"

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{bc: bc}
}

func (bv *BlockValidator) ValidateBlock(b *Block) error {
	if bv.bc.HasBlock(b.Height) {
		return fmt.Errorf("chain already contain this block (%d) with hash (%s)", b.Height, b.Hash(BlockHasher{}))
	}

	// if it's not after last block
	if b.Height != bv.bc.Height()+1 {
		return fmt.Errorf("block (%s)to high", b.Hash(BlockHasher{}))
	}

	prevHeader, err := bv.bc.GetHeader(b.Height - 1)

	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("hash of previous block (%s) is invalid", b.Hash(BlockHasher{}))
	}

	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
