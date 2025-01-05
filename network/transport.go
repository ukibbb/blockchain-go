package network

import "net"

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(net.Addr, []byte) error
	Addr() net.Addr
	Broadcast([]byte) error
}
