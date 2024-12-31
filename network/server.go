package network

import (
	"crypto"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ukibbb/blockchain-go/core"
)

type ServerOpts struct {
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	memPool     *TxPool
	isValidator bool
	blockTime   time.Duration

	rpcCh  chan RPC
	quitch chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		ServerOpts: opts,
		blockTime:  opts.BlockTime,
		memPool:    NewTxPool(),
		// if we are validator we need a private key
		// if it's simple node we are not going to
		// sign blocks
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitch:      make(chan struct{}, 1),
	}
}
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.blockTime)

free:
	for {
		// default here would keep looping
		// and it's cpu intensive.
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%v\n", rpc)
		case <-s.quitch:
			break free
		case <-ticker.C:
			if s.isValidator {
				s.createNewBlock()
			}
		}
	}
}

func (s *Server) handleTransaction(tx *core.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}

	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": tx.Hash(core.TxHasher{}),
		}).Info("transaction already in mempool", hash)
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"hash": hash,
	}).Info("adding new tx to mempool", hash)

	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {

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
