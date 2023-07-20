package network

type GetStatusMessage struct{}

type StatusMessage struct {
	ID            string // the id of the server
	Version       uint32
	CurrentHeight uint32
}
