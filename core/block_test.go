package core

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHeader_Encode_Decode(t *testing.T) {
	hEncode := &Header{
		Version:       1,
		PreviousBlock: types.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        10,
		Nonce:         989394,
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, hEncode.EncodeBinary(buf))

	hDecode := &Header{}
	assert.Nil(t, hDecode.DecodeBinary(buf))

	assert.Equal(t, hEncode, hDecode)
}

func TestBlock_Encode_Decode(t *testing.T) {
	bEncode := &Block{
		Header: Header{
			Version:       1,
			PreviousBlock: types.RandomHash(),
			Timestamp:     time.Now().UnixNano(),
			Height:        10,
			Nonce:         989394,
		},
		Transactions: nil,
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, bEncode.EncodeBinary(buf))

	bDecode := &Block{}
	assert.Nil(t, bDecode.DecodeBinary(buf))

	assert.Equal(t, bEncode, bDecode)
}

func TestBlock_Hash(t *testing.T) {
	b := &Block{
		Header: Header{
			Version:       1,
			PreviousBlock: types.RandomHash(),
			Timestamp:     time.Now().UnixNano(),
			Height:        10,
			Nonce:         989394,
		},
		Transactions: nil,
	}

	h := b.Hash()
	assert.False(t, h.IsZero())
}
