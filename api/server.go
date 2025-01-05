package api

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
	"github.com/ukibbb/blockchain-go/core"
	"github.com/ukibbb/blockchain-go/types"
)

type APIError struct {
	Error string
}

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type Server struct {
	ServerConfig

	bc *core.BlockChain
}

type TxResponse struct {
	TxCount uint
	Hashes  []string
}

type Block struct {
	Hash          string
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Height        uint32
	Timestamp     int64
	Validator     string
	Signature     string

	TxResponse TxResponse
}

func NewServer(cfg ServerConfig, bc *core.BlockChain) *Server {
	return &Server{
		ServerConfig: cfg,
		bc:           bc,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/block/:idOrHash", s.handleGetBlock)

	return e.Start(s.ListenAddr)
}

func (s *Server) handleGetBlock(c echo.Context) error {
	idOrHash := c.Param("idOrHash")

	height, err := strconv.Atoi(idOrHash)
	// If the error is nil we can assume the height of the block is given.
	if err == nil {
		block, err := s.bc.GetBlock(uint32(height))
		if err != nil {
			return c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		}

		return c.JSON(http.StatusOK, intoJSONBlock(block))
	}

	// otherwise assume its the hash
	b, err := hex.DecodeString(idOrHash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
	}

	block, err := s.bc.GetBlockByHash(types.HashFromBytes(b))
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, intoJSONBlock(block))
}

func intoJSONBlock(block *core.Block) Block {
	txResponse := TxResponse{
		TxCount: uint(len(block.Transactions)),
		Hashes:  make([]string, len(block.Transactions)),
	}

	for i := 0; i < int(txResponse.TxCount); i++ {
		txResponse.Hashes[i] = block.Transactions[i].Hash(core.TxHasher{}).String()
	}

	return Block{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		Version:       block.Header.Version,
		Height:        block.Header.Height,
		DataHash:      block.Header.DataHash.String(),
		PrevBlockHash: block.Header.PrevBlockHash.String(),
		Timestamp:     int64(block.Header.Timestamp),
		Validator:     block.Validator.Address().String(),
		Signature:     block.Signature.String(),
		TxResponse:    txResponse,
	}
}
