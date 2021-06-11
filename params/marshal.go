package params

import (
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
)

// Encode provides an IPLD codec encode interface for Params IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXXX when this package is invoked via init.
func Encode(node ipld.Node, w io.Writer) error {
	// 1KiB can be allocated on the stack, and covers most small nodes
	// without having to grow the buffer and cause allocations.
	enc := make([]byte, 0, 1024)

	enc, err := AppendEncode(enc, node)
	if err != nil {
		return err
	}
	_, err = w.Write(enc)
	return err
}

// AppendEncode is like Encode, but it uses a destination buffer directly.
// This means less copying of bytes, and if the destination has enough capacity,
// fewer allocations.
func AppendEncode(enc []byte, inNode ipld.Node) ([]byte, error) {
	hp := new(types.HashedParams)
	if err := EncodeParams(hp, inNode); err != nil {
		return nil, err
	}
	var err error
	enc, err = hp.Marshal()
	return enc, err
}

// EncodeParams is like Encode, but it writes to a tendermint HashedParams object
func EncodeParams(hp *types.HashedParams, inNode ipld.Node) error {
	builder := dagcosmos.Type.HashedParams.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(hp, node); err != nil {
			return err
		}
	}
	return nil
}

var requiredPackFuncs = []func(*types.HashedParams, ipld.Node) error{
	packMaxBytes,
	packMaxGas,
}

func packMaxBytes(hp *types.HashedParams, node ipld.Node) error {
	bmbNode, err := node.LookupByString("BlockMaxBytes")
	if err != nil {
		return err
	}
	bmb, err := bmbNode.AsInt()
	if err != nil {
		return err
	}
	hp.BlockMaxBytes = bmb
	return nil
}

func packMaxGas(hp *types.HashedParams, node ipld.Node) error {
	bmgNode, err := node.LookupByString("BlockMaxGas")
	if err != nil {
		return err
	}
	bmg, err := bmgNode.AsInt()
	if err != nil {
		return err
	}
	hp.BlockMaxGas = bmg
	return nil
}
