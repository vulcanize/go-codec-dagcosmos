package header_tree

import (
	"io"

	"github.com/ipld/go-ipld-prime"

	mt "github.com/vulcanize/go-codec-dagcosmos/merkle_tree"
)

// Decode provides an IPLD codec decode interface for Tendermint PartTree IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
// This simply wraps mt.DecodeTrieNode with the proper multicodec type
func Decode(na ipld.NodeAssembler, in io.Reader) error {
	return mt.DecodeTrieNode(na, in, MultiCodecType)
}

// DecodeBytes is like Decode, but it uses an input buffer directly.
// Decode will grab or read all the bytes from an io.Reader anyway, so this can
// save having to copy the bytes or create a bytes.Buffer.
// This simply wraps mt.DecodeTrieNodeBytes with the proper multicodec type
func DecodeBytes(na ipld.NodeAssembler, src []byte) error {
	return mt.DecodeTrieNodeBytes(na, src, MultiCodecType)
}
