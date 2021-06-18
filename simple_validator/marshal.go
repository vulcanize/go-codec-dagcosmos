package validator

import (
	"io"

	"github.com/ipld/go-ipld-prime"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
)

// Encode provides an IPLD codec encode interface for Tendermint SimpleValidator IPLDs.
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
	val := new(types.SimpleValidator)
	if err := EncodeSimpleValidator(val, inNode); err != nil {
		return nil, err
	}
	var err error
	enc, err = val.Marshal()
	return enc, err
}

// EncodeSimpleValidator is like Encode, but it writes to a tendermint SimpleValidator object
func EncodeSimpleValidator(sv *types.SimpleValidator, inNode ipld.Node) error {
	builder := dagcosmos.Type.Validator.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(sv, node); err != nil {
			return err
		}
	}
	return nil
}

var requiredPackFuncs = []func(*types.SimpleValidator, ipld.Node) error{
	packPubKey,
	packVotingPower,
}

func packPubKey(v *types.SimpleValidator, node ipld.Node) error {
	pkNode, err := node.LookupByString("PubKey")
	if err != nil {
		return err
	}
	pkBytes, err := pkNode.AsBytes()
	if err != nil {
		return err
	}
	tmpk := new(pc.PublicKey)
	if err := tmpk.Unmarshal(pkBytes); err != nil {
		return err
	}
	v.PubKey = tmpk
	return nil
}

func packVotingPower(v *types.SimpleValidator, node ipld.Node) error {
	vpNode, err := node.LookupByString("VotingPower")
	if err != nil {
		return err
	}
	vp, err := vpNode.AsInt()
	if err != nil {
		return err
	}
	v.VotingPower = vp
	return nil
}
