package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-kit/log"
)

type BlockChain struct {
	logger log.Logger
	store  Storage

	mu      sync.RWMutex
	headers []*Header

	validator Validator
}

func NewBlockChain(l log.Logger, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		logger: log.NewLogfmtLogger(os.Stdout),
		store:  NewMemoryStore(),
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
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("givent height (%d) to high", height)
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.headers[height], nil
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.mu.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.mu.Unlock()
	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", b.Transactions,
	)
	return bc.store.Put(b)

}
