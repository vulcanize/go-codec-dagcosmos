package header

import (
	"fmt"
	"io"

	"github.com/vulcanize/go-codec-dagcosmos/shared"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
)

// Encode provides an IPLD codec encode interface for Tendermint header IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
// It expects the root node a HeaderTree (merkle tree of all the header fields)
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
	header := new(types.Header)
	if err := EncodeHeader(header, inNode); err != nil {
		return enc, err
	}
	ph := header.ToProto()
	var err error
	enc, err = ph.Marshal()
	return enc, err
}

// EncodeHeader packs the node into the provided Tendermint Header
func EncodeHeader(header *types.Header, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.Header.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(header, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos Header form (%v)", err)
		}
	}
	return nil
}

var requiredPackFuncs = []func(*types.Header, ipld.Node) error{
	packVersion,
	packChainID,
	packHeight,
	packTime,
	packLastBlockID,
	packLastCommitHash,
	packDataHash,
	packValidatorsHash,
	packNextValidatorsHash,
	packConsensusHash,
	packAppHash,
	packLastResultsHash,
	packEvidenceHash,
	packProposerAddress,
}

func packVersion(h *types.Header, node ipld.Node) error {
	versionNode, err := node.LookupByString("Version")
	if err != nil {
		return err
	}
	version, err := shared.PackVersion(versionNode)
	if err != nil {
		return err
	}
	h.Version = version
	return nil
}

func packChainID(h *types.Header, node ipld.Node) error {
	chainIDNode, err := node.LookupByString("ChainID")
	if err != nil {
		return err
	}
	chainIDStr, err := chainIDNode.AsString()
	if err != nil {
		return err
	}
	h.ChainID = chainIDStr
	return nil
}

func packHeight(h *types.Header, node ipld.Node) error {
	heightNode, err := node.LookupByString("Height")
	if err != nil {
		return err
	}
	heightInt, err := heightNode.AsInt()
	if err != nil {
		return err
	}
	h.Height = heightInt
	return nil
}

func packTime(h *types.Header, node ipld.Node) error {
	timeNode, err := node.LookupByString("Time")
	if err != nil {
		return err
	}
	time, err := shared.PackTime(timeNode)
	if err != nil {
		return err
	}
	h.Time = time
	return nil
}

func packLastBlockID(h *types.Header, node ipld.Node) error {
	lbidNode, err := node.LookupByString("LastBlockID")
	if err != nil {
		return err
	}
	blockID, err := shared.PackBlockID(lbidNode)
	if err != nil {
		return err
	}
	h.LastBlockID = blockID
	return nil
}

func packLastCommitHash(h *types.Header, node ipld.Node) error {
	lastCommitHashNode, err := node.LookupByString("LastCommitHash")
	if err != nil {
		return err
	}
	lastCommitDigest, err := shared.PackLink(lastCommitHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header LastCommitHash multihash: %v", err)
	}
	h.LastCommitHash = lastCommitDigest
	return nil
}

func packDataHash(h *types.Header, node ipld.Node) error {
	dataHashHashNode, err := node.LookupByString("DataHash")
	if err != nil {
		return err
	}
	dataHashDigest, err := shared.PackLink(dataHashHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header DataHash multihash: %v", err)
	}
	h.DataHash = dataHashDigest
	return nil
}

func packValidatorsHash(h *types.Header, node ipld.Node) error {
	validatorHashHashNode, err := node.LookupByString("ValidatorsHash")
	if err != nil {
		return err
	}
	valHashDigest, err := shared.PackLink(validatorHashHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header ValidatorsHash multihash: %v", err)
	}
	h.ValidatorsHash = valHashDigest
	return nil
}

func packNextValidatorsHash(h *types.Header, node ipld.Node) error {
	validatorHashHashNode, err := node.LookupByString("NextValidatorsHash")
	if err != nil {
		return err
	}
	valHashDigest, err := shared.PackLink(validatorHashHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header NextValidatorsHash multihash: %v", err)
	}
	h.NextValidatorsHash = valHashDigest
	return nil
}

func packConsensusHash(h *types.Header, node ipld.Node) error {
	consensusHashHashNode, err := node.LookupByString("ConsensusHash")
	if err != nil {
		return err
	}
	conHashDigest, err := shared.PackLink(consensusHashHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header ConsensusHash multihash: %v", err)
	}
	h.ConsensusHash = conHashDigest
	return nil
}

func packAppHash(h *types.Header, node ipld.Node) error {
	appHashHashNode, err := node.LookupByString("AppHash")
	if err != nil {
		return err
	}
	appHashDigest, err := shared.PackLink(appHashHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header AppHash multihash: %v", err)
	}
	h.AppHash = appHashDigest
	return nil
}

func packLastResultsHash(h *types.Header, node ipld.Node) error {
	lastResultHashNode, err := node.LookupByString("LastResultsHash")
	if err != nil {
		return err
	}
	lastResDigest, err := shared.PackLink(lastResultHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header LastResultsHash multihash: %v", err)
	}
	h.LastResultsHash = lastResDigest
	return nil
}

func packEvidenceHash(h *types.Header, node ipld.Node) error {
	evidenceHashNode, err := node.LookupByString("EvidenceHash")
	if err != nil {
		return err
	}
	evidenceHashDigest, err := shared.PackLink(evidenceHashNode)
	if err != nil {
		return fmt.Errorf("unable to decode header EvidenceHash multihash: %v", err)
	}
	h.EvidenceHash = evidenceHashDigest
	return nil
}

func packProposerAddress(h *types.Header, node ipld.Node) error {
	propserAddrNode, err := node.LookupByString("ProposerAddress")
	if err != nil {
		return err
	}
	propserAddrBytes, err := propserAddrNode.AsBytes()
	if err != nil {
		return err
	}
	h.ProposerAddress = propserAddrBytes
	return nil
}
