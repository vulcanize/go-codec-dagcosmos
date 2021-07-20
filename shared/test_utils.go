package shared

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmBytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/rand"
)

func RandomHash() tmBytes.HexBytes {
	return rand.Bytes(32)
}

func RandomAddr() tmBytes.HexBytes {
	return rand.Bytes(20)
}

func RandomSig() tmBytes.HexBytes {
	return rand.Bytes(ed25519.SignatureSize)
}
