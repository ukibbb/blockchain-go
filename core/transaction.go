package core

import (
	"fmt"

	"github.com/ukibbb/blockchain-go/crypto"
)

type Transaction struct {
	Data []byte

	PublicKey crypto.PublicKey
	Sginature *crypto.Signature
}

func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.PublicKey = privKey.PublicKey()
	tx.Sginature = sig
	return nil
}
func (tx *Transaction) Verify() error {
	if tx.Sginature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	if !tx.Sginature.Verify(tx.PublicKey, tx.Data) {
		return fmt.Errorf("invalida transaction signature")
	}
	return nil

}
