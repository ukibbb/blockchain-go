package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ukibbb/blockchain-go/crypto"
)

func TestSingTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	tx := &Transaction{
		Data: []byte("Transaction data"),
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Sginature)
	assert.Equal(t, tx.From, privKey.PublicKey())
}
func TestVerifyTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	tx := &Transaction{
		Data: []byte("Transaction data"),
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.Nil(t, tx.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()

	assert.NotNil(t, tx.Verify())

}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("foo"),
	}
	assert.Nil(t, tx.Sign(privKey))
	return tx
}
