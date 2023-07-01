package core

import "io"

// Encoder
type Encoder[T any] interface {
	Encode(io.Writer, T) error
}
