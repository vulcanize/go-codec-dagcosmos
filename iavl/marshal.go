package iavl

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

type NodeKind string

var (
	hasher      = sha256.New()
	hashSize    = hasher.Size()
	placeHolder = []byte{} // TODO: it's not clear to me yet what a placeholder hash actually is in the IAVL spec
)

const (
	INNER_NODE NodeKind = "inner"
	LEAF_NODE  NodeKind = "leaf"
)

func (n NodeKind) String() string {
	return string(n)
}

// Encode provides an IPLD codec encode interface for Tendermint IAVL node IPLDs.
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
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.IAVLNode.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return nil, err
	}
	n := builder.Build()
	node, kind, err := NodeAndKind(n)
	if err != nil {
		return nil, err
	}
	switch kind {
	case INNER_NODE:
		enc, err = packInnerNode(node)
	case LEAF_NODE:
		enc, err = packLeafNode(node)
	default:
		return nil, fmt.Errorf("IPLD node is missing the expected Union keys")
	}
	return enc, err
}

func packInnerNode(node ipld.Node) ([]byte, error) {
	var leftData, rightData []byte
	leftNode, err := node.LookupByString("Left")
	if err != nil {
		return nil, err
	}
	if leftNode.IsNull() {
		leftData = placeHolder
	} else {
		leftData, err = shared.PackLink(leftNode)
		if err != nil {
			return nil, err
		}
	}
	rightNode, err := node.LookupByString("Right")
	if err != nil {
		return nil, err
	}
	if rightNode.IsNull() {
		rightData = placeHolder
	} else {
		rightData, err = shared.PackLink(rightNode)
		if err != nil {
			return nil, err
		}
	}
	versionNode, err := node.LookupByString("Version")
	if err != nil {
		return nil, err
	}
	version, err := versionNode.AsInt()
	if err != nil {
		return nil, err
	}
	sizeNode, err := node.LookupByString("Size")
	if err != nil {
		return nil, err
	}
	size, err := sizeNode.AsInt()
	if err != nil {
		return nil, err
	}
	heightNode, err := node.LookupByString("Height")
	if err != nil {
		return nil, err
	}
	height, err := heightNode.AsInt()
	if err != nil {
		return nil, err
	}
	iavlNode := Node{
		leftHash:  leftData,
		rightHash: rightData,
		version:   version,
		height:    int8(height),
		size:      size,
	}
	buf := new(bytes.Buffer)
	iavlNode.writeBytes(buf)
	return buf.Bytes(), nil
}

func packLeafNode(node ipld.Node) ([]byte, error) {
	keyNode, err := node.LookupByString("Key")
	if err != nil {
		return nil, err
	}
	key, err := keyNode.AsBytes()
	if err != nil {
		return nil, err
	}
	valueNode, err := node.LookupByString("Value")
	if err != nil {
		return nil, err
	}
	value, err := valueNode.AsBytes()
	if err != nil {
		return nil, err
	}
	versionNode, err := node.LookupByString("Version")
	if err != nil {
		return nil, err
	}
	version, err := versionNode.AsInt()
	if err != nil {
		return nil, err
	}
	sizeNode, err := node.LookupByString("Size")
	if err != nil {
		return nil, err
	}
	size, err := sizeNode.AsInt()
	if err != nil {
		return nil, err
	}
	heightNode, err := node.LookupByString("Height")
	if err != nil {
		return nil, err
	}
	height, err := heightNode.AsInt()
	if err != nil {
		return nil, err
	}
	iavlNode := Node{
		key:     key,
		value:   value,
		version: version,
		height:  int8(height),
		size:    size,
	}
	buf := new(bytes.Buffer)
	iavlNode.writeBytes(buf)
	return buf.Bytes(), nil
}

// NodeAndKind returns the node and its kind
func NodeAndKind(node ipld.Node) (ipld.Node, NodeKind, error) {
	n, err := node.LookupByString(LEAF_NODE.String())
	if err == nil {
		return n, LEAF_NODE, nil
	}
	n, err = node.LookupByString(INNER_NODE.String())
	if err == nil {
		return n, INNER_NODE, nil
	}
	return nil, "", fmt.Errorf("IAVL IPLD node is missing the expected keyed Union keys")
}

// Node represents a node in a IAVL tree.
type Node struct {
	key       []byte
	value     []byte
	leftHash  []byte
	rightHash []byte
	version   int64
	size      int64
	height    int8
}

// For whatever reason, none of these utils are exported from the original IAVL codebase so we have to duplicate it all
func (node *Node) writeBytes(w io.Writer) error {
	if node == nil {
		return fmt.Errorf("cannot write nil node")
	}
	err := encodeVarint(w, int64(node.height))
	if err != nil {
		return err
	}
	err = encodeVarint(w, node.size)
	if err != nil {
		return err
	}
	err = encodeVarint(w, node.version)
	if err != nil {
		return err
	}

	// Unlike writeHashBytes, key is written for inner nodes.
	err = encodeBytes(w, node.key)
	if err != nil {
		return err
	}

	if node.height == 0 {
		err = encodeBytes(w, node.value)
		if err != nil {
			return err
		}
	} else {
		if node.leftHash == nil {
			panic("node.leftHash was nil in writeBytes")
		}
		err = encodeBytes(w, node.leftHash)
		if err != nil {
			return err
		}

		if node.rightHash == nil {
			panic("node.rightHash was nil in writeBytes")
		}
		err = encodeBytes(w, node.rightHash)
		if err != nil {
			return err
		}
	}
	return nil
}

// encodeBytes writes a varint length-prefixed byte slice to the writer.
func encodeBytes(w io.Writer, bz []byte) error {
	err := encodeUvarint(w, uint64(len(bz)))
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}

// encodeUvarint writes a varint-encoded unsigned integer to an io.Writer.
func encodeUvarint(w io.Writer, u uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], u)
	_, err := w.Write(buf[0:n])
	return err
}

// encodeVarint writes a varint-encoded integer to an io.Writer.
func encodeVarint(w io.Writer, i int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)
	_, err := w.Write(buf[0:n])
	return err
}
