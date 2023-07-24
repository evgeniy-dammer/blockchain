package network

type GetBlocksMessage struct {
	From uint32
	To   uint32 // If To is 0 the maximum blocks will be returned.
}

type GetStatusMessage struct{}

type StatusMessage struct {
	ID            string // the id of the server
	Version       uint32
	CurrentHeight uint32
}
