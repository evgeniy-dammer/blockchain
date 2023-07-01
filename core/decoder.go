package core

import "io"

// Decoder
type Decoder[T any] interface {
	Decode(io.Reader, T) error
}
