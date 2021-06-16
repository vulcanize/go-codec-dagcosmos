package commit

import (
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Encode provides an IPLD codec encode interface for Tendermint CommitSig IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
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
	cs := new(types.CommitSig)
	if err := EncodeCommitSig(cs, inNode); err != nil {
		return enc, err
	}
	ph := cs.ToProto()
	var err error
	enc, err = ph.Marshal()
	return enc, err
}

// EncodeCommitSig is like Encode, but it uses a destination CommitSig struct
func EncodeCommitSig(cs *types.CommitSig, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.Header.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(cs, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos CommitSig form (%v)", err)
		}
	}
	return nil
}

var requiredPackFuncs = []func(*types.CommitSig, ipld.Node) error{
	packBlockIDFlag,
	packValidatorAddress,
	packTimestamp,
	packSignature,
}

func packBlockIDFlag(cs *types.CommitSig, node ipld.Node) error {
	bidFlagNode, err := node.LookupByString("BlockIDFlag")
	if err != nil {
		return err
	}
	bigFlag, err := bidFlagNode.AsInt()
	if err != nil {
		return err
	}
	cs.BlockIDFlag = types.BlockIDFlag(bigFlag)
	return nil
}

func packValidatorAddress(cs *types.CommitSig, node ipld.Node) error {
	valNode, err := node.LookupByString("ValidatorAddress")
	if err != nil {
		return err
	}
	valBytes, err := valNode.AsBytes()
	if err != nil {
		return err
	}
	cs.ValidatorAddress = valBytes
	return nil
}

func packTimestamp(cs *types.CommitSig, node ipld.Node) error {
	timestampNode, err := node.LookupByString("Timestamp")
	if err != nil {
		return err
	}
	time, err := shared.PackTime(timestampNode)
	if err != nil {
		return err
	}
	cs.Timestamp = time
	return nil
}

func packSignature(cs *types.CommitSig, node ipld.Node) error {
	sigNode, err := node.LookupByString("Signature")
	if err != nil {
		return err
	}
	sigBytes, err := sigNode.AsBytes()
	if err != nil {
		return err
	}
	cs.Signature = sigBytes
	return nil
}
