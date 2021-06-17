package mt

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/vulcanize/go-codec-dagcosmos/commit"
	"github.com/vulcanize/go-codec-dagcosmos/evidence"
	"github.com/vulcanize/go-codec-dagcosmos/result"
	validator "github.com/vulcanize/go-codec-dagcosmos/simple_validator"

	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
)

type NodeKind string
type ValueKind string

var (
	hasher       = sha256.New()
	pathSize     = hasher.Size()
	placeholder  = bytes.Repeat([]byte{0}, pathSize)
	defaultValue = []byte{}
	leafPrefix   = []byte{0}
	innerPrefix  = []byte{1}
)

const (
	INNER_NODE NodeKind = "inner"
	LEAF_NODE  NodeKind = "leaf"
	ROOT_NODE  NodeKind = "root"

	UNKNOWN_VALUE   ValueKind = "unknown"
	VALIDATOR_VALUE ValueKind = "validator"
	EVIDENCE_VALUE  ValueKind = "evidence"
	TX_VALUE        ValueKind = "tx"
	COMMIT_VALUE    ValueKind = "commit"
	PART_VALUE      ValueKind = "part"
	RESULT_VALUE    ValueKind = "result"
	HEADER_VALUE    ValueKind = "header"
)

func (n NodeKind) String() string {
	return string(n)
}

func (v ValueKind) String() string {
	return string(v)
}

// Encode provides an IPLD codec encode interface for cosmos merkle tree node IPLDs.
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
	case INNER_NODE, ROOT_NODE:
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
		leftData, err = leftNode.AsBytes()
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
		rightData, err = rightNode.AsBytes()
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
	val, err := packValue(node)
	if err != nil {
		return nil, err
	}
	nodeVal := make([]byte, 0, len(leafPrefix)+len(val))
	nodeVal = append(nodeVal, innerPrefix...)
	nodeVal = append(nodeVal, val...)
	return nodeVal, nil
}

func packValue(node ipld.Node) ([]byte, error) {
	valUnionNode, err := node.LookupByString("Value")
	if err != nil {
		return nil, err
	}
	if valUnionNode.IsNull() {
		return []byte{}, nil
	}
	valNode, valKind, err := ValueAndKind(valUnionNode)
	if err != nil {
		return nil, err
	}
	switch valKind {
	case TX_VALUE:
		txLink, err := valNode.AsLink()
		if err != nil {
			return nil, err
		}
		txCIDLink, ok := txLink.(cidlink.Link)
		if !ok {
			return nil, fmt.Errorf("tx link needs to be a CID")
		}
		txMh := txCIDLink.Hash()
		decodedTxMh, err := multihash.Decode(txMh)
		if err != nil {
			return nil, fmt.Errorf("unable to decode Child multihash: %v", err)
		}
		return decodedTxMh.Digest, nil
	case HEADER_VALUE, PART_VALUE:
		// TODO: figure out how to handle header fields and fragmented blocks better than return raw binary for the values
		// with header fields we know what type to unpack that value on only based on the position of the leaf in the tree
		// with block parts it is impossible to decode the values without collecting them all, concatenating the bytes in the order they
		// appear in the leaf nodes, and unmarshalling the bytes into the Tendermint Block protobuf type since the slice size used to split the
		// protobuf encoding up across the parts is arbitrary and individual fields can be fragmented across separate leaf nodes.
		return valNode.AsBytes()
	case VALIDATOR_VALUE:
		buf := new(bytes.Buffer)
		if err := validator.Encode(valNode, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case RESULT_VALUE:
		buf := new(bytes.Buffer)
		if err := result.Encode(valNode, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case COMMIT_VALUE:
		buf := new(bytes.Buffer)
		if err := commit.Encode(valNode, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case EVIDENCE_VALUE:
		buf := new(bytes.Buffer)
		if err := evidence.Encode(valNode, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("tendermint merkle tree value of unexpected kind %s", valKind.String())
	}
}

func ValueAndKind(node ipld.Node) (ipld.Node, ValueKind, error) {
	n, err := node.LookupByString(TX_VALUE.String())
	if err == nil {
		return n, TX_VALUE, nil
	}
	n, err = node.LookupByString(VALIDATOR_VALUE.String())
	if err == nil {
		return n, VALIDATOR_VALUE, nil
	}
	n, err = node.LookupByString(RESULT_VALUE.String())
	if err == nil {
		return n, RESULT_VALUE, nil
	}
	n, err = node.LookupByString(PART_VALUE.String())
	if err == nil {
		return n, PART_VALUE, nil
	}
	n, err = node.LookupByString(EVIDENCE_VALUE.String())
	if err == nil {
		return n, EVIDENCE_VALUE, nil
	}
	n, err = node.LookupByString(COMMIT_VALUE.String())
	if err == nil {
		return n, COMMIT_VALUE, nil
	}
	n, err = node.LookupByString(HEADER_VALUE.String())
	if err == nil {
		return n, HEADER_VALUE, nil
	}
	return nil, "", fmt.Errorf("merkle tree value IPLD node is missing the expected keyed Union keys")
}

func NodeAndKind(node ipld.Node) (ipld.Node, NodeKind, error) {
	n, err := node.LookupByString(LEAF_NODE.String())
	if err == nil {
		return n, LEAF_NODE, nil
	}
	n, err = node.LookupByString(INNER_NODE.String())
	if err == nil {
		return n, INNER_NODE, nil
	}
	n, err = node.LookupByString(ROOT_NODE.String())
	if err == nil {
		return n, ROOT_NODE, nil
	}
	return nil, "", fmt.Errorf("merkle tree IPLD node is missing the expected keyed Union keys")
}
