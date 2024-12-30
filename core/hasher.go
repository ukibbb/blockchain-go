package core

import (
	"crypto/sha256"

	"github.com/ukibbb/blockchain-go/types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func (BlockHasher) Hash(b *Block) types.Hash {
	h := sha256.Sum256(b.HeaderData())
	return types.Hash(h)

}
