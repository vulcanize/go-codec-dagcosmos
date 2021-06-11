package mt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
)

// DecodeTrieNode provides an IPLD codec decode interface for cosmos merkle tree nodes
// It's not possible to meet the Decode(na ipld.NodeAssembler, in io.Reader) interface
// for a function that supports all trie types (multicodec types), unlike with encoding.
// this is used by Decode functions for each tree type, which are the ones registered to their
// corresponding multicodec
func DecodeTrieNode(na ipld.NodeAssembler, in io.Reader, codec uint64) error {
	var src []byte
	if buf, ok := in.(interface{ Bytes() []byte }); ok {
		src = buf.Bytes()
	} else {
		var err error
		src, err = ioutil.ReadAll(in)
		if err != nil {
			return err
		}
	}
	return DecodeTrieNodeBytes(na, src, codec)
}

func decodeNode(src []byte) ([]byte, NodeKind, error) {
	switch {
	case bytes.HasPrefix(src, innerPrefix):
		return src[1:], INNER_NODE, nil
	case bytes.HasPrefix(src, leafPrefix):
		return src[1:], LEAF_NODE, nil
	default:
		return nil, "", fmt.Errorf("merkle tree node has unrecognized prefix %x", src[0])
	}
}

// DecodeTrieNodeBytes is like DecodeTrieNode, but it uses an input buffer directly.
func DecodeTrieNodeBytes(na ipld.NodeAssembler, src []byte, codec uint64) error {
	nodeBytes, kind, err := decodeNode(src)
	if err != nil {
		return err
	}
	ma, err := na.BeginMap(1)
	if err != nil {
		return err
	}
	switch kind {
	case INNER_NODE:
		if err := ma.AssembleKey().AssignString(INNER_NODE.String()); err != nil {
			return err
		}
		extNodeMA, err := ma.AssembleValue().BeginMap(2)
		if err != nil {
			return err
		}
		if err := unpackInnerNode(extNodeMA, nodeBytes, codec); err != nil {
			return err
		}
		if err := extNodeMA.Finish(); err != nil {
			return err
		}
	case LEAF_NODE:
		if err := ma.AssembleKey().AssignString(LEAF_NODE.String()); err != nil {
			return err
		}
		leafNodeMA, err := ma.AssembleValue().BeginMap(2)
		if err != nil {
			return err
		}
		if err := unpackLeafNode(leafNodeMA, nodeBytes, codec); err != nil {
			return err
		}
		if err := leafNodeMA.Finish(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized NodeKind (%s)", kind.String())
	}

	return ma.Finish()
}

func unpackInnerNode(ma ipld.MapAssembler, nodeBytes []byte, codec uint64) error {
	left, right := nodeBytes[:pathSize], nodeBytes[pathSize:]
	if err := ma.AssembleKey().AssignString("Left"); err != nil {
		return err
	}
	leftCID := sha256ToCid(codec, left)
	leftCIDLink := cidlink.Link{Cid: leftCID}
	if err := ma.AssembleValue().AssignLink(leftCIDLink); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Right"); err != nil {
		return err
	}
	rightCID := sha256ToCid(codec, right)
	rightCIDLink := cidlink.Link{Cid: rightCID}
	return ma.AssembleValue().AssignLink(rightCIDLink)
}

func unpackLeafNode(ma ipld.MapAssembler, valBytes []byte, codec uint64) error {
	if err := ma.AssembleKey().AssignString("Value"); err != nil {
		return err
	}
	valUnionNodeMA, err := ma.AssembleValue().BeginMap(1)
	if err != nil {
		return err
	}
	if err := unpackValue(valUnionNodeMA, valBytes, codec); err != nil {
		return err
	}
	return valUnionNodeMA.Finish()
}

func unpackValue(ma ipld.MapAssembler, val []byte, codec uint64) error {
	switch codec {
	case cid.TendermintTxTree:
		if err := ma.AssembleKey().AssignString(TX_VALUE.String()); err != nil {
			return err
		}
		txCID := sha256ToCid(cid.Raw, val)
		txCIDLink := cidlink.Link{Cid: txCID}
		return ma.AssembleValue().AssignLink(txCIDLink)
	case cid.TendermintValidatorTree:
		if err := ma.AssembleKey().AssignString(VALIDATOR_VALUE.String()); err != nil {
			return err
		}
		return dagcosmos_validator.DecodeBytes(ma.AssembleValue(), val)
	case cid.TendermintPartTree:
		if err := ma.AssembleKey().AssignString(PART_VALUE.String()); err != nil {
			return err
		}
		return dagcosmos_part.DecodeBytes(ma.AssembleValue(), val)
	case cid.TendermintResultTree:
		if err := ma.AssembleKey().AssignString(RESULT_VALUE.String()); err != nil {
			return err
		}
		return dagcosmos_result.DecodeBytes(ma.AssembleValue(), val)
	case cid.TendermintEvidenceTree:
		if err := ma.AssembleKey().AssignString(EVIDENCE_VALUE.String()); err != nil {
			return err
		}
		return dagcosmos_evidence.DecodeBytes(ma.AssembleValue(), val)
	case cid.TendermintCommitTree:
		if err := ma.AssembleKey().AssignString(COMMIT_VALUE.String()); err != nil {
			return err
		}
		return dagcosmos_commit.DecodeBytes(ma.AssembleValue(), val)
	case cid.TendermintHeaderTree:
		if err := ma.AssembleKey().AssignString(HEADER_VALUE.String()); err != nil {
			return err
		}
		return ma.AssembleValue().AssignBytes(val)
	default:
		return fmt.Errorf("unsupported multicodec type (%d) for tendermint merkle tree unmarshaller", codec)
	}
}

// sha256ToCid takes a sha256 hash and returns its cid based on
// the codec given.
func sha256ToCid(codec uint64, h []byte) cid.Cid {
	buf, err := multihash.Encode(h, multihash.SHA2_256)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(codec, multihash.Multihash(buf))
}
