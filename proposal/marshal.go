package proposal

import (
	"fmt"
	"io"

	"github.com/tendermint/tendermint/libs/protoio"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/ipld/go-ipld-prime"
	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Encode provides an IPLD codec encode interface for Tendermint Proposal IPLDs.
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
	cp := new(tmproto.CanonicalProposal)
	if err := EncodeCanonicalProposal(cp, inNode); err != nil {
		return enc, err
	}
	var err error
	enc, err = protoio.MarshalDelimited(cp)
	return enc, err
}

// EncodeCanonicalProposal is like Encode, but it uses a destination CanonicalProposal protobuf type
func EncodeCanonicalProposal(cs *tmproto.CanonicalProposal, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.Header.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(cs, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos Proposal form (%v)", err)
		}
	}
	return nil
}

var requiredPackFuncs = []func(*tmproto.CanonicalProposal, ipld.Node) error{
	packType,
	packHeight,
	packRound,
	packPOLRound,
	packBlockID,
	packTimestamp,
	packChainID,
}

func packType(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	typeNode, err := node.LookupByString("Type")
	if err != nil {
		return err
	}
	typeInt, err := typeNode.AsInt()
	if err != nil {
		return err
	}
	cp.Type = tmproto.SignedMsgType(typeInt)
	return nil
}

func packHeight(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	heightNode, err := node.LookupByString("Height")
	if err != nil {
		return err
	}
	heightInt, err := heightNode.AsInt()
	if err != nil {
		return err
	}
	cp.Height = heightInt
	return nil
}

func packRound(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	roundNode, err := node.LookupByString("Round")
	if err != nil {
		return err
	}
	roundInt, err := roundNode.AsInt()
	if err != nil {
		return err
	}
	cp.Round = roundInt
	return nil
}

func packPOLRound(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	polRoundNode, err := node.LookupByString("POLRound")
	if err != nil {
		return err
	}
	polRoundInt, err := polRoundNode.AsInt()
	if err != nil {
		return err
	}
	cp.POLRound = polRoundInt
	return nil
}

func packBlockID(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	blockIDNode, err := node.LookupByString("BlockID")
	if err != nil {
		return err
	}
	blockID, err := shared.PackBlockID(blockIDNode)
	if err != nil {
		return err
	}
	cp.BlockID = &tmproto.CanonicalBlockID{
		Hash: blockID.Hash,
		PartSetHeader: tmproto.CanonicalPartSetHeader{
			Hash:  blockID.PartSetHeader.Hash,
			Total: blockID.PartSetHeader.Total,
		},
	}
	return nil
}

func packTimestamp(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	timestampNode, err := node.LookupByString("Timestamp")
	if err != nil {
		return err
	}
	time, err := shared.PackTime(timestampNode)
	if err != nil {
		return err
	}
	cp.Timestamp = time
	return nil
}

func packChainID(cp *tmproto.CanonicalProposal, node ipld.Node) error {
	chainIDNode, err := node.LookupByString("ChainID")
	if err != nil {
		return err
	}
	chainID, err := chainIDNode.AsString()
	if err != nil {
		return err
	}
	cp.ChainID = chainID
	return nil
}
