package smt

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

// Decode provides an IPLD codec decode interface for Tendermint SMT nodes
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
func Decode(na ipld.NodeAssembler, in io.Reader) error {
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
	return DecodeBytes(na, src)
}

// DecodeBytes is like DecodeTrieNode, but it uses an input buffer directly.
// Decode will grab or read all the bytes from an io.Reader anyway, so this can
// save having to copy the bytes or create a bytes.Buffer.
func DecodeBytes(na ipld.NodeAssembler, src []byte) error {
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
		if err := unpackInnerNode(extNodeMA, nodeBytes); err != nil {
			return err
		}
		if err := extNodeMA.Finish(); err != nil {
			return err
		}
	case LEAF_NODE:
		if err := ma.AssembleKey().AssignString(LEAF_NODE.String()); err != nil {
			return err
		}
		leafNodeMA, err := ma.AssembleValue().BeginMap(1)
		if err != nil {
			return err
		}
		if err := unpackLeafNode(leafNodeMA, nodeBytes); err != nil {
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

func unpackInnerNode(ma ipld.MapAssembler, nodeBytes []byte) error {
	left, right := nodeBytes[:hashSize], nodeBytes[hashSize:]
	if err := ma.AssembleKey().AssignString("Left"); err != nil {
		return err
	}
	if bytes.Equal(left, placeholder) {
		if err := ma.AssembleValue().AssignNull(); err != nil {
			return err
		}
	} else {
		leftCID := sha256ToCid(MultiCodecType, left)
		leftCIDLink := cidlink.Link{Cid: leftCID}
		if err := ma.AssembleValue().AssignLink(leftCIDLink); err != nil {
			return err
		}
	}
	if err := ma.AssembleKey().AssignString("Right"); err != nil {
		return err
	}
	if bytes.Equal(right, placeholder) {
		return ma.AssembleValue().AssignNull()
	}
	rightCID := sha256ToCid(MultiCodecType, right)
	rightCIDLink := cidlink.Link{Cid: rightCID}
	return ma.AssembleValue().AssignLink(rightCIDLink)
}

func unpackLeafNode(ma ipld.MapAssembler, leafBytes []byte) error {
	path, val := leafBytes[:hashSize], leafBytes[hashSize:]
	if err := ma.AssembleKey().AssignString("Value"); err != nil {
		return err
	}
	if err := ma.AssembleValue().AssignBytes(val); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Path"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(path)
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

// returns node without prefix and the kind of the node
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
