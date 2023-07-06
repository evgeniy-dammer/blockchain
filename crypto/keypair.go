package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
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

	return &Signature{R: r, S: s}, nil
}

// GeneratePrivateKey generates a new PrivateKey
func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return PrivateKey{key: key}
}

// PublicKey returns a PublicKey for PrivateKey
func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{Key: &k.key.PublicKey}
}

// PublicKey
type PublicKey struct {
	Key *ecdsa.PublicKey
}

// ToSlice returns PublicKey as a slice of bytes
func (k PublicKey) ToSlice() []byte {
	return elliptic.MarshalCompressed(k.Key, k.Key.X, k.Key.Y)
}

// Address returns an Address of a PublicKey
func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.ToSlice())

	return types.AddressFromBytes(h[len(h)-20:])
}

// Signature
type Signature struct {
	R, S *big.Int
}

// Verify verifies a given slice of bytes with a PublicKey
func (s Signature) Verify(pubKey PublicKey, data []byte) bool {
	return ecdsa.Verify(pubKey.Key, data, s.R, s.S)
}
