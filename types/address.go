package types

import (
	"encoding/hex"
	"fmt"
)

// Address
type Address [20]uint8

// ToSlice returns Address as a slice of bytes
func (a Address) ToSlice() []byte {
	b := make([]byte, 20)

	for i := 0; i < 20; i++ {
		b[i] = a[i]
	}

	return b
}

// String returns Address as a string
func (a Address) String() string {
	return hex.EncodeToString(a.ToSlice())
}

// AddressFromBytes returns an Address from a given slice of bytes
func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		msg := fmt.Sprintf("bytes should be with length 20")
		panic(msg)
	}

	var value [20]uint8

	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}

	return Address(value)
}
