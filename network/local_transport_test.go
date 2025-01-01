package network

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Equal(t, tra.peers[trb.addr], trb)
	assert.Equal(t, trb.peers[tra.addr], tra)

	// assert.Equal(t, 1, 1)

}

func TestSendMessage(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)
	msg := []byte("HelloWorld")

	assert.Nil(t, tra.SendMessage(trb.Addr(), msg))

	rpc := <-trb.Consume()
	b, err := io.ReadAll(rpc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)
	assert.Equal(t, rpc.From, tra.Addr())

}

func TestBroadcast(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")
	trc := NewLocalTransport("C")

	tra.Connect(trb)
	tra.Connect(trc)

	msg := []byte("fooo")
	assert.Nil(t, tra.Broadcast(msg))

	rpcb := <-trb.Consume()
	b, err := io.ReadAll(rpcb.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)

	rpcc := <-trc.Consume()
	c, err := io.ReadAll(rpcc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, c, msg)

}
