package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeypair_Sign_Verify_Valid(t *testing.T) {
	privateKey := GeneratePrivateKey()
	pubKey := privateKey.PublicKey()

	address := pubKey.Address()

	msg := []byte("Helloworld")

	sig, err := privateKey.Sign(msg)

	assert.Nil(t, err)

	fmt.Println("address", address)

	verified := sig.Verify(pubKey, msg)
	assert.True(t, verified)
}

func TestKeypair_Sign_Verify_Fail(t *testing.T) {
	privateKey := GeneratePrivateKey()
	pubKey := privateKey.PublicKey()

	address := pubKey.Address()

	newPrivateKey := GeneratePrivateKey()
	newPublicKey := newPrivateKey.PublicKey()

	msg := []byte("Helloworld")
	sig, err := privateKey.Sign(msg)

	assert.Nil(t, err)

	fmt.Println("address", address)

	verified := sig.Verify(newPublicKey, msg)
	assert.False(t, verified)
	assert.False(t, sig.Verify(pubKey, []byte("differentmsg")))
}
