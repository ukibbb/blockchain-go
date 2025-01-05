package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/ukibbb/blockchain-go/core"
	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	SeedNodes     []string
	ListenAddr    string
	TCPTransport  *TCPTransport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCPRocessor  RPCPRocessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	SeedNodes    []string
	TCPTransport *TCPTransport
	peerMap      map[net.Addr]*TCPPeer

	mu     sync.RWMutex
	peerCh chan *TCPPeer

	ServerOpts
	memPool     *TxPool
	chain       *core.BlockChain
	isValidator bool

	rpcCh  chan RPC
	quitch chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stdout)
		opts.Logger = log.With(opts.Logger, "addr", opts.ID)
	}
	chain, err := core.NewBlockChain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransport(opts.ListenAddr, peerCh)
	s := &Server{
		SeedNodes:    opts.SeedNodes,
		TCPTransport: tr,
		peerCh:       make(chan *TCPPeer),
		peerMap:      make(map[net.Addr]*TCPPeer),
		ServerOpts:   opts,
		chain:        chain,
		memPool:      NewTxPool(1000),
		// if we are validator we need a private key
		// if it's simple node we are not going to
		// sign blocks
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitch:      make(chan struct{}, 1),
	}
	// if not server process from server options, server is processor.
	if s.RPCPRocessor == nil {
		s.RPCPRocessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}
func (s *Server) Start() {
	s.TCPTransport.Start()

	s.bootstrapNework()

free:
	for {
		// default here would keep looping
		// and it's cpu intensive.
		select {
		case peer := <-s.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer

			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}

			s.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, peer.conn.RemoteAddr())

		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
				continue
			}
			if err := s.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("error", err)

				}
			}
		case <-s.quitch:
			break free
		}
	}
}
func (s *Server) bootstrapNework() {
	for _, addr := range s.SeedNodes {
		fmt.Println("trying to connect to ", addr)
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("could not connect to %+v\n", conn)
				return
			}
			s.peerCh <- &TCPPeer{
				conn: conn,
			}

		}(addr)

	}

}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)
	s.Logger.Log(
		"msg", "starting validator loop",
		"blocktime", s.BlockTime,
	)
	for {
		<-ticker.C
		s.createNewBlock()
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		s.processGetStatusMessage(msg.From)
	case *StatusMessage:
		s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		s.processBlocksMessage(msg.From, t)
	}

	return nil
}
func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	s.Logger.Log("msg", "received BLOCKS", "from", from)
	for _, block := range data.Blocks {
		if err := s.chain.AddBlock(block); err != nil {
			fmt.Printf("adding block error %s\n", err)
			continue
		}
	}
	return nil
}

func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	s.Logger.Log("msg", "recieved getBlocks message", "from", from)

	var (
		blocks    = []*core.Block{}
		ourHeight = s.chain.Height()
	)

	if data.To == 0 {
		for i := int(data.From); i <= int(ourHeight); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	return peer.Send(msg.Bytes())
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	s.Logger.Log("msg", "received STATUS message ", "from", from)
	if data.CurrentHeight <= s.chain.Height() {
		s.Logger.Log(
			"msg",
			"cannoct sync block height to low",
			"outHeight", s.chain.Height(),
			"thier height", data.CurrentHeight,
			"addr", from,
		)
		return nil
	}
	// in this case i am sure that node has blocks higher than this
	// var (
	// 	getBlocksMessage = &GetBlocksMessage{
	// 		From: s.chain.Height() + 1, // causse height is genesis block
	// 		To:   0,
	// 	}
	// 	buf = new(bytes.Buffer)
	// )

	// if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
	// 	return err
	// }

	// msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())

	// s.mu.RLock()
	// defer s.mu.RUnlock()
	// peer, ok := s.peerMap[from]
	// if !ok {
	// 	return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	// }

	// return peer.Send(msg.Bytes())
	go s.requestBlocksLoop(from)
	return nil

}
func (s *Server) processGetStatusMessage(from net.Addr) error {
	s.Logger.Log("msg", "received getStatus message", "from", from)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return peer.Send(msg.Bytes())

}
func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			fmt.Printf("send peer error => addr %s [err: %s]\n", netAddr, err)
		}

	}
	return nil
}

func (s *Server) requestBlocksLoop(peer net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)
	for {
		ourHeight := s.chain.Height()
		s.Logger.Log("msg", "requesting new blocks", "currentHeight", ourHeight)
		getBlocksMessage := &GetBlocksMessage{
			From: ourHeight + 1,
			To:   0,
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
			return err
		}

		s.mu.RLock()

		defer s.mu.RUnlock()

		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		peer, ok := s.peerMap[peer]

		if !ok {
			return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
		}
		if err := peer.Send(msg.Bytes()); err != nil {
			s.Logger.Log("error", "failed send to peer", "err", err, "peer", peer)
		}

		<-ticker.C
	}
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}
	go s.broadcastBlock(b)
	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := new(bytes.Buffer)
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	// s.Logger.Log(
	// 	"msg", "adding new tx mem",
	// 	"hash", hash,
	// 	"mempoolLength", s.memPool.PendingCount(),
	// )

	go s.broadcastTx(tx)

	s.memPool.Add(tx)
	return nil
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// for now we are putting everything in block
	// in reality we can put everything
	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)

	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.memPool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	h := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000000,
	}
	b, _ := core.NewBlock(h, nil)

	privKey := crypto.GeneratePrivateKey()
	if err := b.Sign(privKey); err != nil {
		panic(err)
	}

	return b
}
