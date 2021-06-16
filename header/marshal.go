package header

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/vulcanize/go-codec-dagcosmos/shared"

	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
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

/*
return to this later

// EncodeHeader packs the node into the provided go-ethereum Header
func EncodeHeader(header *types.Header, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.MerkleTreeInnerNode.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	// node needs to be the root node of the header tree
	_, err := node.LookupByString("root")
	if err != nil {
		return fmt.Errorf("the ipld node provided tendermint header Encode functions must be the root node of a header field merkle tree")
	}
	for _, pFunc := range requiredPackFuncs {
		if err := pFunc(header, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos Header form (%v)", err)
		}
	}
	return nil
}

func NodeAndKind(node ipld.Node) (ipld.Node, mt.NodeKind, error) {
	n, err := node.LookupByString(mt.INNER_NODE.String())
	if err == nil {
		return n, mt.INNER_NODE, nil
	}
	n, err = node.LookupByString(mt.ROOT_NODE.String())
	if err == nil {
		return n, mt.ROOT_NODE, nil
	}
	n, err = node.LookupByString(mt.LEAF_NODE.String())
	if err == nil {
		return n, mt.LEAF_NODE, nil
	}
	return nil, "", fmt.Errorf("unrecognized merke tree node kind")
}

func CollectLeafNodes(node ipld.Node) ([]ipld.Node, error) {
	n, kind, err := NodeAndKind(node)
	if err != nil {
		return nil, err
	}
	switch kind {
	case mt.INNER_NODE, mt.ROOT_NODE:
		leftNode, err := n.LookupByString("Left")
		if err != nil {
			return nil, err
		}
		l, err := leftNode.AsLink()
		if err != nil {
			return nil, err
		}
	case mt.LEAF_NODE:
	default:
		return nil, fmt.Errorf("unrecognized merke tree node kind")
	}

}
*/
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
	packProsperAddress,
}

func packVersion(h *types.Header, node ipld.Node) error {
	versionNode, err := node.LookupByString("Version")
	if err != nil {
		return err
	}
	blockVersionNode, err := versionNode.LookupByString("Block")
	if err != nil {
		return err
	}
	blockVersionBytes, err := blockVersionNode.AsBytes()
	if err != nil {
		return err
	}
	appVersionNode, err := versionNode.LookupByString("App")
	if err != nil {
		return err
	}
	appVersionBytes, err := appVersionNode.AsBytes()
	if err != nil {
		return err
	}
	h.Version = tmversion.Consensus{
		Block: binary.BigEndian.Uint64(blockVersionBytes),
		App:   binary.BigEndian.Uint64(appVersionBytes),
	}
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
	lastCommitLink, err := lastCommitHashNode.AsLink()
	if err != nil {
		return err
	}
	lastCommitCIDLink, ok := lastCommitLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a LastCommitHash link")
	}
	lcMh := lastCommitCIDLink.Hash()
	decodedLcMh, err := multihash.Decode(lcMh)
	if err != nil {
		return fmt.Errorf("unable to decode header LastCommitHash multihash: %v", err)
	}
	h.LastCommitHash = decodedLcMh.Digest
	return nil
}

func packDataHash(h *types.Header, node ipld.Node) error {
	dataHashHashNode, err := node.LookupByString("DataHash")
	if err != nil {
		return err
	}
	dataHashLink, err := dataHashHashNode.AsLink()
	if err != nil {
		return err
	}
	dataHashCIDLink, ok := dataHashLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a DataHash link")
	}
	dhMh := dataHashCIDLink.Hash()
	decodedDhMh, err := multihash.Decode(dhMh)
	if err != nil {
		return fmt.Errorf("unable to decode header DataHash multihash: %v", err)
	}
	h.DataHash = decodedDhMh.Digest
	return nil
}

func packValidatorsHash(h *types.Header, node ipld.Node) error {
	validatorHashHashNode, err := node.LookupByString("ValidatorsHash")
	if err != nil {
		return err
	}
	validatorHashLink, err := validatorHashHashNode.AsLink()
	if err != nil {
		return err
	}
	validatorHashCIDLink, ok := validatorHashLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a ValidatorsHash link")
	}
	vhMh := validatorHashCIDLink.Hash()
	decodedVhMh, err := multihash.Decode(vhMh)
	if err != nil {
		return fmt.Errorf("unable to decode header ValidatorsHash multihash: %v", err)
	}
	h.ValidatorsHash = decodedVhMh.Digest
	return nil
}

func packNextValidatorsHash(h *types.Header, node ipld.Node) error {
	validatorHashHashNode, err := node.LookupByString("NextValidatorsHash")
	if err != nil {
		return err
	}
	validatorHashLink, err := validatorHashHashNode.AsLink()
	if err != nil {
		return err
	}
	validatorHashCIDLink, ok := validatorHashLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a NextValidatorsHash link")
	}
	vhMh := validatorHashCIDLink.Hash()
	decodedVhMh, err := multihash.Decode(vhMh)
	if err != nil {
		return fmt.Errorf("unable to decode header NextValidatorsHash multihash: %v", err)
	}
	h.NextValidatorsHash = decodedVhMh.Digest
	return nil
}

func packConsensusHash(h *types.Header, node ipld.Node) error {
	consensusHashHashNode, err := node.LookupByString("ConsensusHash")
	if err != nil {
		return err
	}
	consensusHashLink, err := consensusHashHashNode.AsLink()
	if err != nil {
		return err
	}
	consensusHashCIDLink, ok := consensusHashLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a ConsensusHash link")
	}
	chMh := consensusHashCIDLink.Hash()
	decodedChMh, err := multihash.Decode(chMh)
	if err != nil {
		return fmt.Errorf("unable to decode header ConsensusHash multihash: %v", err)
	}
	h.ConsensusHash = decodedChMh.Digest
	return nil
}

func packAppHash(h *types.Header, node ipld.Node) error {
	appHashHashNode, err := node.LookupByString("AppHash")
	if err != nil {
		return err
	}
	appHashLink, err := appHashHashNode.AsLink()
	if err != nil {
		return err
	}
	appHashCIDLink, ok := appHashLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a AppHash link")
	}
	ahMh := appHashCIDLink.Hash()
	decodedAhMh, err := multihash.Decode(ahMh)
	if err != nil {
		return fmt.Errorf("unable to decode header AppHash multihash: %v", err)
	}
	h.AppHash = decodedAhMh.Digest
	return nil
}

func packLastResultsHash(h *types.Header, node ipld.Node) error {
	lastResultHashNode, err := node.LookupByString("LastResultsHash")
	if err != nil {
		return err
	}
	lastResultLink, err := lastResultHashNode.AsLink()
	if err != nil {
		return err
	}
	lastResultCIDLink, ok := lastResultLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a LastResultsHash link")
	}
	lhMh := lastResultCIDLink.Hash()
	decodedLhMh, err := multihash.Decode(lhMh)
	if err != nil {
		return fmt.Errorf("unable to decode header LastResultsHash multihash: %v", err)
	}
	h.LastResultsHash = decodedLhMh.Digest
	return nil
}

func packEvidenceHash(h *types.Header, node ipld.Node) error {
	evidenceHashNode, err := node.LookupByString("EvidenceHash")
	if err != nil {
		return err
	}
	evidenceLink, err := evidenceHashNode.AsLink()
	if err != nil {
		return err
	}
	evidenceCIDLink, ok := evidenceLink.(cidlink.Link)
	if !ok {
		return fmt.Errorf("header must have a EvidenceHash link")
	}
	eMh := evidenceCIDLink.Hash()
	decodedEMh, err := multihash.Decode(eMh)
	if err != nil {
		return fmt.Errorf("unable to decode header EvidenceHash multihash: %v", err)
	}
	h.EvidenceHash = decodedEMh.Digest
	return nil
}

func packProsperAddress(h *types.Header, node ipld.Node) error {
	propserAddrNode, err := node.LookupByString("ProsperAddress")
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
