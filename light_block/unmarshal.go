package light_block

import (
	"encoding/binary"
	"io"
	"io/ioutil"

	gogotypes "github.com/gogo/protobuf/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/tendermint/tendermint/types"
)

// Decode provides an IPLD codec decode interface for Cosmos LightBlock IPLDs.
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

// DecodeBytes is like Decode, but it uses an input buffer directly.
// Decode will grab or read all the bytes from an io.Reader anyway, so this can
// save having to copy the bytes or create a bytes.Buffer.
func DecodeBytes(na ipld.NodeAssembler, src []byte) error {
	tmlb := new(tmproto.LightBlock)
	if err := tmlb.Unmarshal(src); err != nil {
		return err
	}
	lb, err := types.LightBlockFromProto(tmlb)
	if err != nil {
		return err
	}
	return DecodeLightBlock(na, *lb)
}

// DecodeLightBlock is like Decode, but it uses an input tendermint LightBlock type
func DecodeLightBlock(na ipld.NodeAssembler, lb types.LightBlock) error {
	ma, err := na.BeginMap(15)
	if err != nil {
		return err
	}
	for _, upFunc := range requiredUnpackFuncs {
		if err := upFunc(ma, lb); err != nil {
			return err
		}
	}
	return ma.Finish()
}

var requiredUnpackFuncs = []func(ipld.MapAssembler, types.LightBlock) error{
	unpackVersion,
	unpackChainID,
	unpackHeight,
	unpackTime,
	unpackLastBlockID,
	unpackLastCommitHash,
	unpackDataHash,
	unpackValidatorsHash,
	unpackNextValidatorsHash,
	unpackConsensusHash,
	unpackAppHash,
	unpackLastResultsHash,
	unpackEvidenceHash,
	unpackProsperAddress,
}

func unpackVersion(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("Version"); err != nil {
		return err
	}
	vma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("Block"); err != nil {
		return err
	}
	blockVerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockVerBytes, h.Version.Block)
	if err := vma.AssembleValue().AssignBytes(blockVerBytes); err != nil {
		return err
	}
	if err := vma.AssembleKey().AssignString("App"); err != nil {
		return err
	}
	appVerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(appVerBytes, h.Version.App)
	if err := vma.AssembleValue().AssignBytes(appVerBytes); err != nil {
		return err
	}
	return vma.Finish()
}

func unpackChainID(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("ChainID"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignString(h.ChainID)
}

func unpackHeight(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(h.Height)
}

func unpackTime(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("Time"); err != nil {
		return err
	}
	tma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	timestamp, err := gogotypes.TimestampProto(h.Time)
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

func unpackLastBlockID(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("LastBlockID"); err != nil {
		return err
	}
	lbaMa, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := lbaMa.AssembleKey().AssignString("Hash"); err != nil {
		return err
	}
	headerMh, err := multihash.Encode(h.LastBlockId.Hash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use HeaderTree codec type?
	headerCID := cid.NewCidV1(MultiCodecType, headerMh)
	headerLinkCID := cidlink.Link{Cid: headerCID}
	if err := lbaMa.AssembleValue().AssignLink(headerLinkCID); err != nil {
		return err
	}
	if err := lbaMa.AssembleKey().AssignString("PartSetHeader"); err != nil {
		return err
	}
	pshMa, err := lbaMa.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := pshMa.AssembleKey().AssignString("Total"); err != nil {
		return err
	}
	totalBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(totalBytes, h.LastBlockId.PartSetHeader.Total)
	if err := pshMa.AssembleValue().AssignBytes(totalBytes); err != nil {
		return err
	}
	if err := pshMa.AssembleKey().AssignString("Hash"); err != nil {
		return err
	}
	partMh, err := multihash.Encode(h.LastBlockId.PartSetHeader.Hash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use PartTree codec type
	partCID := cid.NewCidV1(MultiCodecType, partMh)
	partLinkCID := cidlink.Link{Cid: partCID}
	if err := pshMa.AssembleValue().AssignLink(partLinkCID); err != nil {
		return err
	}
	if err := pshMa.Finish(); err != nil {
		return err
	}
	return lbaMa.Finish()
}

func unpackLastCommitHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("LastCommitHash"); err != nil {
		return err
	}
	lchMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use CommitTree codec type
	lcCID := cid.NewCidV1(MultiCodecType, lchMh)
	lcLinkCID := cidlink.Link{Cid: lcCID}
	return ma.AssembleValue().AssignLink(lcLinkCID)
}

func unpackDataHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("DataHash"); err != nil {
		return err
	}
	dataMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use TxTree codec type
	dataCID := cid.NewCidV1(MultiCodecType, dataMh)
	dataLinkCID := cidlink.Link{Cid: dataCID}
	return ma.AssembleValue().AssignLink(dataLinkCID)
}

func unpackValidatorsHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("ValidatorsHash"); err != nil {
		return err
	}
	valMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use ValidatorTree codec type
	valCID := cid.NewCidV1(MultiCodecType, valMh)
	valLinkCID := cidlink.Link{Cid: valCID}
	return ma.AssembleValue().AssignLink(valLinkCID)
}

func unpackNextValidatorsHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("NextValidatorsHash"); err != nil {
		return err
	}
	valMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use ValidatorTree codec type
	valCID := cid.NewCidV1(MultiCodecType, valMh)
	valLinkCID := cidlink.Link{Cid: valCID}
	return ma.AssembleValue().AssignLink(valLinkCID)
}

func unpackConsensusHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("ConsensusHash"); err != nil {
		return err
	}
	conMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use HashedParams codec type
	conCID := cid.NewCidV1(MultiCodecType, conMh)
	conLinkCID := cidlink.Link{Cid: conCID}
	return ma.AssembleValue().AssignLink(conLinkCID)
}

func unpackAppHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("AppHash"); err != nil {
		return err
	}
	appMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use AppTree (IAVL or SMT) codec type
	appCID := cid.NewCidV1(MultiCodecType, appMh)
	appLinkCID := cidlink.Link{Cid: appCID}
	return ma.AssembleValue().AssignLink(appLinkCID)
}

func unpackLastResultsHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("LastResultsHash"); err != nil {
		return err
	}
	lrMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use ResultTree codec type
	lrCID := cid.NewCidV1(MultiCodecType, lrMh)
	lrLinkCID := cidlink.Link{Cid: lrCID}
	return ma.AssembleValue().AssignLink(lrLinkCID)
}

func unpackEvidenceHash(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("EvidenceHash"); err != nil {
		return err
	}
	eMh, err := multihash.Encode(h.LastCommitHash, MultiHashType)
	if err != nil {
		return err
	}
	// TODO: switch to use EvidenceTree codec type
	eCID := cid.NewCidV1(MultiCodecType, eMh)
	eLinkCID := cidlink.Link{Cid: eCID}
	return ma.AssembleValue().AssignLink(eLinkCID)
}

func unpackProsperAddress(ma ipld.MapAssembler, h types.Header) error {
	if err := ma.AssembleKey().AssignString("ProsperAddress"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(h.ProposerAddress)
}
