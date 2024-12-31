package core

import (
	"fmt"

	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

type Transaction struct {
	Data []byte

	// public key of transaction is
	// public key of the sender
	From      crypto.PublicKey
	Sginature *crypto.Signature

	// cached version of tx data
	hash types.Hash
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
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
