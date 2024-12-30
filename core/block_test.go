package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

func randomBlock(height uint32) *Block {
	h := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     uint64(time.Now().UnixNano()),
		Height:        height,
		Nonce:         987654567,
	}
	data := []byte("fooo")
	tx := Transaction{Data: data}

	return &Block{Header: h, Transactions: []Transaction{tx}}

}

func TestSignBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)

}
func TestVerifyBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()

	assert.NotNil(t, b.Verify())

	// alter data so this is not same block
	b.Height = 100
	assert.NotNil(t, b.Verify())

}
