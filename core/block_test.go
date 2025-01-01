package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

func TestSignBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)

}
func TestVerifyBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()

	assert.NotNil(t, b.Verify())

	// alter data so this is not same block
	b.Height = 100
	assert.NotNil(t, b.Verify())

}

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	h := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     uint64(time.Now().UnixNano()),
		Height:        height,
		Nonce:         987654567,
	}

	b, _ := NewBlock(h, []Transaction{tx})
	dataHash, err := CalculateDataHash([]Transaction{tx})

	assert.Nil(t, err)

	b.Header.DataHash = dataHash

	assert.Nil(t, b.Sign(privKey))
	return b
}
