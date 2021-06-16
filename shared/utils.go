package shared

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/vulcanize/go-codec-dagcosmos/commit"

	"github.com/tendermint/tendermint/crypto/encoding"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/bytes"

	"github.com/ipld/go-ipld-prime"
)

func GetTxType(node ipld.Node) (uint8, error) {
	tyNode, err := node.LookupByString("Type")
	if err != nil {
		return 0, err
	}
	tyBytes, err := tyNode.AsBytes()
	if err != nil {
		return 0, err
	}
	if len(tyBytes) != 1 {
		return 0, fmt.Errorf("tx type should be a single byte")
	}
	return tyBytes[0], nil
}

type WriteableByteSlice []byte

func (w *WriteableByteSlice) Write(b []byte) (int, error) {
	*w = append(*w, b...)
	return len(b), nil
}

// PackBlockID returns the blockID from the provided ipld.Node
func PackBlockID(node ipld.Node) (types.BlockID, error) {
	headerHashNode, err := node.LookupByString("Hash")
	if err != nil {
		return types.BlockID{}, err
	}
	headerLink, err := headerHashNode.AsLink()
	if err != nil {
		return types.BlockID{}, err
	}
	headerCIDLink, ok := headerLink.(cidlink.Link)
	if !ok {
		return types.BlockID{}, fmt.Errorf("header must have a Hash link")
	}
	headerMh := headerCIDLink.Hash()
	decodedHeaderMh, err := multihash.Decode(headerMh)
	if err != nil {
		return types.BlockID{}, fmt.Errorf("unable to decode header Hash multihash: %v", err)
	}

	partSetHeaderNode, err := node.LookupByString("PartSetHeader")
	if err != nil {
		return types.BlockID{}, err
	}
	totalNode, err := partSetHeaderNode.LookupByString("Total")
	if err != nil {
		return types.BlockID{}, err
	}
	totalBytes, err := totalNode.AsBytes()
	if err != nil {
		return types.BlockID{}, err
	}

	partHashNode, err := partSetHeaderNode.LookupByString("Hash")
	if err != nil {
		return types.BlockID{}, err
	}
	partTreeLink, err := partHashNode.AsLink()
	if err != nil {
		return types.BlockID{}, err
	}
	partTreeCIDLink, ok := partTreeLink.(cidlink.Link)
	if !ok {
		return types.BlockID{}, fmt.Errorf("header PartSetHeader must have a Hash link")
	}
	partTreeMh := partTreeCIDLink.Hash()
	decodedPartTreeMh, err := multihash.Decode(partTreeMh)
	if err != nil {
		return types.BlockID{}, fmt.Errorf("unable to decode header PartSetHeader Hash multihash: %v", err)
	}

	return types.BlockID{
		Hash: decodedHeaderMh.Digest,
		PartSetHeader: types.PartSetHeader{
			Total: binary.BigEndian.Uint32(totalBytes),
			Hash:  decodedPartTreeMh.Digest,
		},
	}, nil
}

// PackValidator packs a Validator from the provided ipld.Node
func PackValidator(validatorNode ipld.Node) (*types.Validator, error) {
	addrNode, err := validatorNode.LookupByString("Address")
	if err != nil {
		return nil, err
	}
	addr, err := addrNode.AsBytes()
	if err != nil {
		return nil, err
	}
	pkNode, err := validatorNode.LookupByString("PubKey")
	if err != nil {
		return nil, err
	}
	pkBytes, err := pkNode.AsBytes()
	if err != nil {
		return nil, err
	}
	tmpk := new(pc.PublicKey)
	if err := tmpk.Unmarshal(pkBytes); err != nil {
		return nil, err
	}
	pk, err := encoding.PubKeyFromProto(*tmpk)
	if err != nil {
		return nil, err
	}
	vpNode, err := validatorNode.LookupByString("VotingPower")
	if err != nil {
		return nil, err
	}
	vp, err := vpNode.AsInt()
	if err != nil {
		return nil, err
	}
	ppNode, err := validatorNode.LookupByString("ProposerPriority")
	if err != nil {
		return nil, err
	}
	pp, err := ppNode.AsInt()
	if err != nil {
		return nil, err
	}
	return &types.Validator{
		Address:          addr,
		PubKey:           pk,
		VotingPower:      vp,
		ProposerPriority: pp,
	}, nil
}

// PackCommit packs a Commit from the provided ipld.Node
func PackCommit(commitNode ipld.Node) (*types.Commit, error) {
	heightNode, err := commitNode.LookupByString("Height")
	if err != nil {
		return nil, err
	}
	height, err := heightNode.AsInt()
	if err != nil {
		return nil, err
	}
	roundNode, err := commitNode.LookupByString("Round")
	if err != nil {
		return nil, err
	}
	round, err := roundNode.AsInt()
	if err != nil {
		return nil, err
	}
	blockIDNode, err := commitNode.LookupByString("BlockID")
	blockID, err := PackBlockID(blockIDNode)
	if err != nil {
		return nil, err
	}
	signaturesNode, err := commitNode.LookupByString("Signatures")
	if err != nil {
		return nil, err
	}
	signatures := make([]types.CommitSig, signaturesNode.Length())
	signatureLI := signaturesNode.ListIterator()
	for !signatureLI.Done() {
		i, commitSigNode, err := signatureLI.Next()
		if err != nil {
			return nil, err
		}
		commitSig := new(types.CommitSig)
		if err := commit.EncodeCommitSig(commitSig, commitSigNode); err != nil {
			return nil, err
		}
		signatures[i] = *commitSig
	}
	return &types.Commit{
		Height:     height,
		Round:      int32(round),
		BlockID:    blockID,
		Signatures: signatures,
	}, nil
}

// PackTime returns the timestamp from the provided ipld.Node
func PackTime(timeNode ipld.Node) (time.Time, error) {
	secondsNode, err := timeNode.LookupByString("Seconds")
	if err != nil {
		return time.Time{}, err
	}
	seconds, err := secondsNode.AsInt()
	if err != nil {
		return time.Time{}, err
	}
	nanoSecondsNode, err := timeNode.LookupByString("Nanoseconds")
	if err != nil {
		return time.Time{}, err
	}
	nanoSeconds, err := nanoSecondsNode.AsInt()
	if err != nil {
		return time.Time{}, err
	}
	timestamp := &gogotypes.Timestamp{
		Seconds: seconds,
		Nanos:   int32(nanoSeconds),
	}
	return gogotypes.TimestampFromProto(timestamp)
}

// PackVote returns the Vote from the provided ipld.Node
func PackVote(voteNode ipld.Node) (*types.Vote, error) {
	vote := new(types.Vote)
	voteTypeNode, err := voteNode.LookupByString("Type")
	if err != nil {
		return nil, nil
	}
	voteType, err := voteTypeNode.AsInt()
	if err != nil {
		return nil, nil
	}
	vote.Type = tmproto.SignedMsgType(voteType)

	heightNode, err := voteNode.LookupByString("Height")
	if err != nil {
		return nil, nil
	}
	height, err := heightNode.AsInt()
	if err != nil {
		return nil, nil
	}
	vote.Height = height

	roundNode, err := voteNode.LookupByString("Round")
	if err != nil {
		return nil, nil
	}
	round, err := roundNode.AsInt()
	if err != nil {
		return nil, nil
	}
	vote.Round = int32(round)

	bidNode, err := voteNode.LookupByString("BlockID")
	if err != nil {
		return nil, nil
	}
	blockID, err := PackBlockID(bidNode)
	if err != nil {
		return nil, nil
	}
	vote.BlockID = blockID

	timeNode, err := voteNode.LookupByString("Timestamp")
	if err != nil {
		return nil, nil
	}
	time, err := PackTime(timeNode)
	if err != nil {
		return nil, nil
	}
	vote.Timestamp = time

	valNode, err := voteNode.LookupByString("ValidatorAddress")
	if err != nil {
		return nil, nil
	}
	val, err := valNode.AsBytes()
	if err != nil {
		return nil, nil
	}
	vote.ValidatorAddress = val

	iNode, err := voteNode.LookupByString("ValidatorIndex")
	if err != nil {
		return nil, nil
	}
	index, err := iNode.AsInt()
	if err != nil {
		return nil, nil
	}
	vote.ValidatorIndex = int32(index)

	sigNode, err := voteNode.LookupByString("Signature")
	if err != nil {
		return nil, nil
	}
	sig, err := sigNode.AsBytes()
	if err != nil {
		return nil, nil
	}
	vote.Signature = sig
	return vote, nil
}

// CdcEncode returns nil if the input is nil, otherwise returns
// proto.Marshal(<type>Value{Value: item})
func CdcEncode(item interface{}) []byte {
	if item != nil && !IsTypedNil(item) && !IsEmpty(item) {
		switch item := item.(type) {
		case string:
			i := gogotypes.StringValue{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		case int64:
			i := gogotypes.Int64Value{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		case bytes.HexBytes:
			i := gogotypes.BytesValue{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		default:
			return nil
		}
	}

	return nil
}

// IsTypedNil return true if a value is nil
func IsTypedNil(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// IsEmpty returns true if it has zero length.
func IsEmpty(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	default:
		return false
	}
}
