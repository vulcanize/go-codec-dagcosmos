package smt

import (
	"bytes"
	"crypto/sha256"
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
	placeholder = bytes.Repeat([]byte{0}, hashSize)
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

const (
	INNER_NODE NodeKind = "inner"
	LEAF_NODE  NodeKind = "leaf"
)

func (n NodeKind) String() string {
	return string(n)
}

// Encode provides an IPLD codec encode interface for Tendermint SMT node IPLDs.
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
	builder := dagcosmos.Type.MerkleTreeNode.NewBuilder()
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
		leftData = placeholder
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
		rightData = placeholder
	} else {
		rightData, err = shared.PackLink(rightNode)
		if err != nil {
			return nil, err
		}
	}
	nodeVal := make([]byte, 0, len(innerPrefix)+len(leftData)+len(rightData))
	nodeVal = append(nodeVal, innerPrefix...)
	nodeVal = append(nodeVal, leftData...)
	nodeVal = append(nodeVal, rightData...)
	return nodeVal, nil
}

func packLeafNode(node ipld.Node) ([]byte, error) {
	path, err := packPath(node)
	if err != nil {
		return nil, err
	}
	val, err := packValue(node)
	if err != nil {
		return nil, err
	}
	nodeVal := make([]byte, 0, len(leafPrefix)+len(path)+len(val))
	nodeVal = append(nodeVal, leafPrefix...)
	nodeVal = append(nodeVal, path...)
	nodeVal = append(nodeVal, val...)
	return nodeVal, nil
}

func packPath(node ipld.Node) ([]byte, error) {
	pathNode, err := node.LookupByString("Path")
	if err != nil {
		return nil, err
	}
	return pathNode.AsBytes()
}

func packValue(node ipld.Node) ([]byte, error) {
	valNode, err := node.LookupByString("Value")
	if err != nil {
		return nil, err
	}
	// at this time we are not attempting to further handle the arbitrary hash(key, value) stored here
	return valNode.AsBytes()
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
	return nil, "", fmt.Errorf("SMT IPLD node is missing the expected keyed Union keys")
}
