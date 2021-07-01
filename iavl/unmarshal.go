package iavl

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
)

// Decode provides an IPLD codec decode interface for Tendermint IAVL nodes
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
	iavlNode, kind, err := decodeNode(src)
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
		if err := unpackInnerNode(extNodeMA, *iavlNode); err != nil {
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
		if err := unpackLeafNode(leafNodeMA, *iavlNode); err != nil {
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

func unpackInnerNode(ma ipld.MapAssembler, iavlNode Node) error {
	if err := ma.AssembleKey().AssignString("Left"); err != nil {
		return err
	}
	if bytes.Equal(iavlNode.leftHash, placeHolder) {
		if err := ma.AssembleValue().AssignNull(); err != nil {
			return err
		}
	} else {
		leftCID := sha256ToCid(MultiCodecType, iavlNode.leftHash)
		leftCIDLink := cidlink.Link{Cid: leftCID}
		if err := ma.AssembleValue().AssignLink(leftCIDLink); err != nil {
			return err
		}
	}
	if err := ma.AssembleKey().AssignString("Right"); err != nil {
		return err
	}
	if bytes.Equal(iavlNode.rightHash, placeHolder) {
		if err := ma.AssembleValue().AssignNull(); err != nil {
			return err
		}
	} else {
		rightCID := sha256ToCid(MultiCodecType, iavlNode.rightHash)
		rightCIDLink := cidlink.Link{Cid: rightCID}
		if err := ma.AssembleValue().AssignLink(rightCIDLink); err != nil {
			return err
		}
	}
	if err := ma.AssembleKey().AssignString("Version"); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignInt(iavlNode.version); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignInt(int64(iavlNode.height)); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Size"); err != nil {
		return err
	}
	return ma.AssembleKey().AssignInt(iavlNode.size)
}

func unpackLeafNode(ma ipld.MapAssembler, iavlNode Node) error {
	if err := ma.AssembleKey().AssignString("Value"); err != nil {
		return err
	}
	if err := ma.AssembleValue().AssignBytes(iavlNode.value); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Key"); err != nil {
		return err
	}
	if err := ma.AssembleValue().AssignBytes(iavlNode.key); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Version"); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignInt(iavlNode.version); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignInt(int64(iavlNode.height)); err != nil {
		return err
	}
	if err := ma.AssembleKey().AssignString("Size"); err != nil {
		return err
	}
	return ma.AssembleKey().AssignInt(iavlNode.size)
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
func decodeNode(src []byte) (*Node, NodeKind, error) {
	node, err := makeNode(src)
	if err != nil {
		return nil, "", err
	}
	if node.height == 0 {
		return node, LEAF_NODE, nil
	}
	return node, INNER_NODE, nil
}

// In the original implementation there is an exported function for decoding nodes but it returns the iavl.Node type that doesn't export any of its fields...
func makeNode(buf []byte) (*Node, error) {

	// Read node header (height, size, version, key).
	height, n, err := decodeVarint(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]
	if height < int64(math.MinInt8) || height > int64(math.MaxInt8) {
		return nil, errors.New("invalid height, must be int8")
	}

	size, n, err := decodeVarint(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	ver, n, err := decodeVarint(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	key, n, err := decodeBytes(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	node := &Node{
		height:  int8(height),
		size:    size,
		version: ver,
		key:     key,
	}

	// Read node body.

	if node.height == 0 {
		val, _, err := decodeBytes(buf)
		if err != nil {
			return nil, err
		}
		node.value = val
	} else { // Read children.
		leftHash, n, err := decodeBytes(buf)
		if err != nil {
			return nil, err
		}
		buf = buf[n:]

		rightHash, _, err := decodeBytes(buf)
		if err != nil {
			return nil, err
		}
		node.leftHash = leftHash
		node.rightHash = rightHash
	}
	return node, nil
}

// decodeBytes decodes a varint length-prefixed byte slice, returning it along with the number
// of input bytes read.
func decodeBytes(bz []byte) ([]byte, int, error) {
	s, n, err := decodeUvarint(bz)
	if err != nil {
		return nil, n, err
	}
	// Make sure size doesn't overflow. ^uint(0) >> 1 will help determine the
	// max int value variably on 32-bit and 64-bit machines. We also doublecheck
	// that size is positive.
	size := int(s)
	if s >= uint64(^uint(0)>>1) || size < 0 {
		return nil, n, fmt.Errorf("invalid out of range length %v decoding []byte", s)
	}
	// Make sure end index doesn't overflow. We know n>0 from decodeUvarint().
	end := n + size
	if end < n {
		return nil, n, fmt.Errorf("invalid out of range length %v decoding []byte", size)
	}
	// Make sure the end index is within bounds.
	if len(bz) < end {
		return nil, n, fmt.Errorf("insufficient bytes decoding []byte of length %v", size)
	}
	bz2 := make([]byte, size)
	copy(bz2, bz[n:end])
	return bz2, end, nil
}

// decodeUvarint decodes a varint-encoded unsigned integer from a byte slice, returning it and the
// number of bytes decoded.
func decodeUvarint(bz []byte) (uint64, int, error) {
	u, n := binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		return u, n, errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		return u, n, errors.New("EOF decoding uvarint")
	}
	return u, n, nil
}

// decodeVarint decodes a varint-encoded integer from a byte slice, returning it and the number of
// bytes decoded.
func decodeVarint(bz []byte) (int64, int, error) {
	i, n := binary.Varint(bz)
	if n == 0 {
		return i, n, errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		return i, n, errors.New("EOF decoding varint")
	}
	return i, n, nil
}
