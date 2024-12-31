package core

import (
	"fmt"

	"github.com/ukibbb/blockchain-go/crypto"
)

type Transaction struct {
	Data []byte

	// public key of transaction is
	// public key of the sender
	From      crypto.PublicKey
	Sginature *crypto.Signature
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
