package result

import (
	"encoding/binary"
	"io"

	"github.com/ipld/go-ipld-prime"
	abci "github.com/tendermint/tendermint/abci/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
)

// Encode provides an IPLD codec encode interface for Tendermint Result IPLDs.
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
	res := new(abci.ResponseDeliverTx)
	if err := EncodeResult(res, inNode); err != nil {
		return nil, err
	}
	var err error
	enc, err = res.Marshal()
	return enc, err
}

// EncodeResult is like Encode, but it writes to a tendermint ResponseDeliverTx object
func EncodeResult(hp *abci.ResponseDeliverTx, inNode ipld.Node) error {
	builder := dagcosmos.Type.ResponseDeliverTx.NewBuilder()
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

var requiredPackFuncs = []func(*abci.ResponseDeliverTx, ipld.Node) error{
	packCode,
	packData,
	packGasWanted,
	packGasUsed,
}

func packCode(res *abci.ResponseDeliverTx, node ipld.Node) error {
	codeNode, err := node.LookupByString("Code")
	if err != nil {
		return err
	}
	codeBytes, err := codeNode.AsBytes()
	if err != nil {
		return err
	}
	res.Code = binary.BigEndian.Uint32(codeBytes)
	return nil
}

func packData(res *abci.ResponseDeliverTx, node ipld.Node) error {
	dataNode, err := node.LookupByString("Data")
	if err != nil {
		return err
	}
	dataBytes, err := dataNode.AsBytes()
	if err != nil {
		return err
	}
	res.Data = dataBytes
	return nil
}

func packGasWanted(res *abci.ResponseDeliverTx, node ipld.Node) error {
	gwNode, err := node.LookupByString("GasWanted")
	if err != nil {
		return err
	}
	gw, err := gwNode.AsInt()
	if err != nil {
		return err
	}
	res.GasWanted = gw
	return nil
}

func packGasUsed(res *abci.ResponseDeliverTx, node ipld.Node) error {
	guNode, err := node.LookupByString("GasUsed")
	if err != nil {
		return err
	}
	gu, err := guNode.AsInt()
	if err != nil {
		return err
	}
	res.GasUsed = gu
	return nil
}
