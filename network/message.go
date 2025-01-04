package network

type GetBlocksMessage struct {
	// from this height to that height
	From uint32
	// if To is 0 maximum block will be returned
	To uint32
}

type GetStatusMessage struct{}

type StatusMessage struct {
	ID string
	// Version       uint32
	CurrentHeight uint32
}
