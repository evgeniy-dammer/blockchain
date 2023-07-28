package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/evgeniy-dammer/blockchain/types"
	"math/big"
)

// PrivateKey
type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// Sign returns a Signature of a given slice of bytes
func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}

	return &Signature{
		R: r,
		S: s,
	}, nil
}

// GeneratePrivateKey generates a new PrivateKey
func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return PrivateKey{
		key: key,
	}
}

// PublicKey returns a PublicKey for PrivateKey
func (k PrivateKey) PublicKey() PublicKey {
	return elliptic.MarshalCompressed(k.key.PublicKey, k.key.PublicKey.X, k.key.PublicKey.Y)
}

// PublicKey
type PublicKey []byte

// Address returns an Address of a PublicKey
func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k)

	return types.AddressFromBytes(h[len(h)-20:])
}

// Signature
type Signature struct {
	S *big.Int
	R *big.Int
}

func (sig Signature) String() string {
	b := append(sig.S.Bytes(), sig.R.Bytes()...)
	return hex.EncodeToString(b)
}

// Verify verifies a given slice of bytes with a PublicKey
func (s Signature) Verify(pubKey PublicKey, data []byte) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubKey)
	key := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(key, data, s.R, s.S)
}
