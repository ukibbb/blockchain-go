package core

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"

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

func (b *Block) Sign(prvKey crypto.PrivateKey) error {
	sig, err := prvKey.Sign(b.HeaderData())
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
	if !b.Signature.Verify(b.Validator, b.HeaderData()) {
		return errors.New("block has invalid  sginatur")
	}
	return nil
}

func (b *Block) Hash(hasher Hasher[*Block]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}
	return b.hash
}

func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) HeaderData() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(b.Header)
	return buf.Bytes()
}
