package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Hash
type Hash [32]uint8

// IsZero checks if Hash is zero
func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}

	return true
}

// ToSlice returns Hash as a slice of bytes
func (h Hash) ToSlice() []byte {
	b := make([]byte, 32)

	for i := 0; i < 32; i++ {
		b[i] = h[i]
	}

	return b
}

// String returns Hash as a string
func (h Hash) String() string {
	return hex.EncodeToString(h.ToSlice())
}

// HashFromBytes hashes a given slice of bytes
func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		msg := fmt.Sprintf("bytes should be with length 32")
		panic(msg)
	}

	var value [32]uint8

	for i := 0; i < 32; i++ {
		value[i] = b[i]
	}

	return Hash(value)
}

// RandomBytes returns a random slice of bytes
func RandomBytes(size int) []byte {
	token := make([]byte, size)
	_, err := rand.Read(token)
	if err != nil {
		return nil
	}

	return token
}

// RandomHash returns a random Hash
func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}
