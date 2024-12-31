package core

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     uint64
	Height        uint32
	Nonce         uint64
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(h)
	return buf.Bytes()

}

type Block struct {
	*Header
	Transactions []Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature

	// Cached version of the header hash
	hash types.Hash
}

func NewBlock(h *Header, txx []Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txx,
	}
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, *tx)

}

func (b *Block) Sign(prvKey crypto.PrivateKey) error {
	sig, err := prvKey.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}
	b.Validator = prvKey.PublicKey()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return errors.New("block has no signature")
	}
	if !b.Signature.Verify(b.Validator, b.Header.Bytes()) {
		return errors.New("block has invalid  sginatur")
	}

	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}
	return nil
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}
