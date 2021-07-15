package shared

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/tendermint/tendermint/crypto/encoding"
	"github.com/tendermint/tendermint/libs/bytes"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

// PackLink returns the hash digest from a link
func PackLink(node ipld.Node) ([]byte, error) {
	dl, err := node.AsLink()
	if err != nil {
		return nil, err
	}
	dcl, ok := dl.(cidlink.Link)
	if !ok {
		return nil, fmt.Errorf("unable to decode Merkle tree node multihash %v", err)
	}
	dmh := dcl.Hash()
	ddmh, err := multihash.Decode(dmh)
	if err != nil {
		return nil, fmt.Errorf("unable to decode Merkle tree node multihash: %v", err)
	}
	return ddmh.Digest, nil
}

// PackBlockID returns the blockID from the provided ipld.Node
func PackBlockID(node ipld.Node) (types.BlockID, error) {
	headerHashNode, err := node.LookupByString("Hash")
	if err != nil {
		return types.BlockID{}, err
	}
	headerDigest, err := PackLink(headerHashNode)
	if err != nil {
		return types.BlockID{}, err
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
	partTreeDigest, err := PackLink(partHashNode)
	if err != nil {
		return types.BlockID{}, err
	}

	return types.BlockID{
		Hash: headerDigest,
		PartSetHeader: types.PartSetHeader{
			Total: binary.BigEndian.Uint32(totalBytes),
			Hash:  partTreeDigest,
		},
	}, nil
}

// UnpackBlockID unpacks BlockID into MapAssembler
func UnpackBlockID(bima ipld.MapAssembler, bid types.BlockID) error {
	if err := bima.AssembleKey().AssignString("Hash"); err != nil {
		return err
	}
	headerMh, err := multihash.Encode(bid.Hash, multihash.SHA2_256)
	if err != nil {
		return err
	}
	// TODO: switch to use HeaderTree codec type?
	headerCID := cid.NewCidV1(cid.DagCBOR, headerMh)
	headerLinkCID := cidlink.Link{Cid: headerCID}
	if err := bima.AssembleValue().AssignLink(headerLinkCID); err != nil {
		return err
	}
	if err := bima.AssembleKey().AssignString("PartSetHeader"); err != nil {
		return err
	}
	pshMa, err := bima.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := pshMa.AssembleKey().AssignString("Total"); err != nil {
		return err
	}
	totalBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(totalBytes, bid.PartSetHeader.Total)
	if err := pshMa.AssembleValue().AssignBytes(totalBytes); err != nil {
		return err
	}
	if err := pshMa.AssembleKey().AssignString("Hash"); err != nil {
		return err
	}
	partMh, err := multihash.Encode(bid.PartSetHeader.Hash, multihash.SHA2_256)
	if err != nil {
		return err
	}
	// TODO: switch to use PartTree codec type
	partCID := cid.NewCidV1(cid.DagCBOR, partMh)
	partLinkCID := cidlink.Link{Cid: partCID}
	if err := pshMa.AssembleValue().AssignLink(partLinkCID); err != nil {
		return err
	}
	if err := pshMa.Finish(); err != nil {
		return err
	}
	return bima.Finish()
}

// PackValidator packs a Validator from the provided ipld.Node
// This includes all fields of the Validator, not just the consensus fields
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

// UnpackValidator unpacks Validator into MapAssembler
// This includes all fields on the Validator, not just the consensus fields
func UnpackValidator(vama ipld.MapAssembler, validator types.Validator) error {
	if err := vama.AssembleKey().AssignString("Address"); err != nil {
		return err
	}
	if err := vama.AssembleValue().AssignBytes(validator.Address); err != nil {
		return err
	}
	if err := vama.AssembleKey().AssignString("PubKey"); err != nil {
		return err
	}
	tmpk, err := encoding.PubKeyToProto(validator.PubKey)
	if err != nil {
		return err
	}
	tmpkBytes, err := tmpk.Marshal()
	if err != nil {
		return err
	}
	if err := vama.AssembleValue().AssignBytes(tmpkBytes); err != nil {
		return err
	}
	if err := vama.AssembleKey().AssignString("VotingPower"); err != nil {
		return err
	}
	if err := vama.AssembleValue().AssignInt(validator.VotingPower); err != nil {
		return err
	}
	if err := vama.AssembleKey().AssignString("ProposerPriority"); err != nil {
		return err
	}
	if err := vama.AssembleValue().AssignInt(validator.ProposerPriority); err != nil {
		return err
	}
	return vama.Finish()
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

// UnpackTime unpacks the provided time into the MapAssembler
func UnpackTime(tma ipld.MapAssembler, t time.Time) error {
	timestamp, err := gogotypes.TimestampProto(t)
	if err != nil {
		return err
	}
	if err := tma.AssembleKey().AssignString("Seconds"); err != nil {
		return err
	}
	if err := tma.AssembleValue().AssignInt(timestamp.Seconds); err != nil {
		return err
	}
	if err := tma.AssembleKey().AssignString("Nanoseconds"); err != nil {
		return err
	}
	if err := tma.AssembleValue().AssignInt(int64(timestamp.Nanos)); err != nil {
		return err
	}
	return tma.Finish()
}

// PackVote returns the Vote from the provided ipld.Node
func PackVote(voteNode ipld.Node) (*types.Vote, error) {
	vote := new(types.Vote)
	voteTypeNode, err := voteNode.LookupByString("SMType")
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

// UnpackVote unpacks Vote into MapAssembler
func UnpackVote(vma ipld.MapAssembler, vote types.Vote) error {
	if err := vma.AssembleKey().AssignString("SMType"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignInt(int64(vote.Type)); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignInt(vote.Height); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("Round"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignInt(int64(vote.Round)); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("BlockID"); err != nil {
		return err
	}
	biMA, err := vma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := UnpackBlockID(biMA, vote.BlockID); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("Timestamp"); err != nil {
		return err
	}
	tiMA, err := vma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := UnpackTime(tiMA, vote.Timestamp); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("ValidatorAddress"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignBytes(vote.ValidatorAddress); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("ValidatorIndex"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignInt(int64(vote.ValidatorIndex)); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("Signature"); err != nil {
		return err
	}
	if err := vma.AssembleValue().AssignBytes(vote.Signature); err != nil {
		return err
	}
	return vma.Finish()
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
