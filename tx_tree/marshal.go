package tx_tree

import (
	"io"

	"github.com/ipld/go-ipld-prime"

	mt "github.com/vulcanize/go-codec-dagcosmos/merkle_tree"
)

// Encode provides an IPLD codec encode interface for Tendermint TxTree IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
// This is a pure wrapping around mt.Encode to expose it from this package
func Encode(node ipld.Node, w io.Writer) error {
	return mt.Encode(node, w)
}

// AppendEncode is like Encode, but it uses a destination buffer directly.
// This means less copying of bytes, and if the destination has enough capacity,
// fewer allocations.
// This is a pure wrapping around mt.AppendEncode to expose it from this package
func AppendEncode(enc []byte, inNode ipld.Node) ([]byte, error) {
	return mt.AppendEncode(enc, inNode)
}
