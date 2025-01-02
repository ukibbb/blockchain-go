package network

import (
	"bytes"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/ukibbb/blockchain-go/core"
	"github.com/ukibbb/blockchain-go/crypto"
	"github.com/ukibbb/blockchain-go/types"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCPRocessor  RPCPRocessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
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
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	chain, err := core.NewBlockChain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOpts: opts,
		chain:      chain,
		memPool:    NewTxPool(),
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
	s.initTransports()

free:
	for {
		// default here would keep looping
		// and it's cpu intensive.
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}
			if err := s.ProcessMessage(msg); err != nil {
				s.Logger.Log("error", err)
			}
		case <-s.quitch:
			break free
		}
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
	}
	return nil
}

func (s *Server) broadcast(msg []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	return nil
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
	if s.memPool.Has(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	s.Logger.Log(
		"msg", "adding new tx mem",
		"hash", hash,
		"mempoolLength", s.memPool.Len(),
	)

	go s.broadcastTx(tx)

	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// for now we are putting everything in block
	// in reality we can put everything
	txx := s.memPool.Transactions()

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

	s.memPool.Flush()

	return nil
}

func (s *Server) initTransports() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}

func genesisBlock() *core.Block {
	h := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000000,
	}
	b, _ := core.NewBlock(h, nil)
	return b
}
