package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeypair_SignVerifySuccess(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	msg := []byte("Hello World!")

	sign, err := privKey.Sign(msg)
	assert.Nil(t, err)
	assert.True(t, sign.Verify(pubKey, msg))
}

func TestKeypair_SignVerifyFail(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	msg := []byte("Hello World!")

	sign, err := privKey.Sign(msg)
	assert.Nil(t, err)

	otherPrivKey := GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()

	assert.False(t, sign.Verify(otherPubKey, msg))
	assert.False(t, sign.Verify(pubKey, []byte("World Hello!")))
}
