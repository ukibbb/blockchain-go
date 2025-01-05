package main

import (
	"bytes"
	"encoding/gob"
	"net"
	"time"

	"github.com/ukibbb/blockchain-go/core"
	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/network"
)

func main() {

	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":3000"})
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":5000"})
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", nil)
	go remoteNodeB.Start()

	go func() {
		time.Sleep(20 * time.Second)
		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"})
		go lateNode.Start()
	}()

	time.Sleep(1 * time.Second)
	select {}
}

func sendGetStatusMessage(tr network.Transport, to net.Addr) error {
	var (
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())

}

func sendTransaction(tr network.Transport, to net.Addr) error {
	// helper function

	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}

func makeServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string) *network.Server {
	opts := network.ServerOpts{
		SeedNodes:  seedNodes,
		ListenAddr: addr,
		PrivateKey: pk,
		ID:         id,
	}

	s, err := network.NewServer(opts)
	if err != nil {
	}
	return s
}
