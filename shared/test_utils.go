package shared

import (
	tmBytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/rand"
)

func RandomHash() tmBytes.HexBytes {
	return rand.Bytes(32)
}

func RandomAddr() tmBytes.HexBytes {
	return rand.Bytes(20)
}
