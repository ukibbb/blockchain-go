package core

import (
	"encoding/gob"
	"fmt"
	"math/rand"

	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

type TxType byte

const (
	TxTypeCollection TxType = iota
	TxTypeMint
)

type CollectionTx struct {
	Fee      int64
	MetaData []byte
}

type MintTx struct {
	Fee        int64
	NFT        types.Hash
	Collection types.Hash

	MetaData []byte

	CollectionOwner crypto.PublicKey
	Signature       crypto.Signature
}

type Transaction struct {
	Type    TxType
	TxInner any
	Data    []byte

	// public key of transaction is
	// public key of the sender
	From      crypto.PublicKey
	Sginature *crypto.Signature

	// cached version of tx data
	hash types.Hash

	Nonce uint64

	// firstSeen is timestamp of when this ts is seen locally
	firstSeen int64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: uint64(rand.Int63n(1000000000000000)),
	}
}

func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.From = privKey.PublicKey()
	tx.Sginature = sig
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Sginature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	if !tx.Sginature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalida transaction signature")
	}
	return nil
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) SetFirstSeen(t int64) {
	tx.firstSeen = t
}

func (tx *Transaction) FirstSeen() int64 {
	return tx.firstSeen
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}
func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func init() {
	gob.Register(CollectionTx{})
	gob.Register(MintTx{})
}
