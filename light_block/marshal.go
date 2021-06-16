package light_block

import (
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/header"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Encode provides an IPLD codec encode interface for Tendermint light block IPLDs.
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
	lb := new(types.LightBlock)
	if err := EncodeLightBlock(lb, inNode); err != nil {
		return enc, err
	}
	tmlb, err := lb.ToProto()
	if err != nil {
		return nil, err
	}
	enc, err = tmlb.Marshal()
	return enc, err
}

// EncodeLightBlock packs the node into the provided Tendermint LightBlock
func EncodeLightBlock(lb *types.LightBlock, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.LightBlock.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(lb, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos Header form (%v)", err)
		}
	}
	return nil
}

var requiredPackFuncs = []func(*types.LightBlock, ipld.Node) error{
	packSignedHeader,
	packValidatorSet,
}

func packSignedHeader(lb *types.LightBlock, node ipld.Node) error {
	shNode, err := node.LookupByString("SignedHeader")
	if err != nil {
		return err
	}
	hNode, err := shNode.LookupByString("Header")
	if err != nil {
		return err
	}
	h := new(types.Header)
	if err := header.EncodeHeader(h, hNode); err != nil {
		return err
	}
	lb.SignedHeader.Header = h
	commitNode, err := shNode.LookupByString("Commit")
	if err != nil {
		return err
	}
	commit, err := shared.PackCommit(commitNode)
	if err != nil {
		return err
	}
	lb.SignedHeader.Commit = commit
	return nil
}

func packValidatorSet(lb *types.LightBlock, node ipld.Node) error {
	vSetNode, err := node.LookupByString("ValidatorSet")
	if err != nil {
		return err
	}
	validatorsNode, err := vSetNode.LookupByString("Validators")
	if err != nil {
		return err
	}
	validators := make([]*types.Validator, validatorsNode.Length())
	valsIT := validatorsNode.ListIterator()
	for !valsIT.Done() {
		i, validatorNode, err := valsIT.Next()
		if err != nil {
			return err
		}
		validator, err := shared.PackValidator(validatorNode)
		if err != nil {
			return err
		}
		validators[i] = validator
	}
	lb.ValidatorSet.Validators = validators
	proposerNode, err := vSetNode.LookupByString("Proposer")
	if err != nil {
		return err
	}
	proposer, err := shared.PackValidator(proposerNode)
	if err != nil {
		return err
	}
	lb.ValidatorSet.Proposer = proposer
	return nil
}
