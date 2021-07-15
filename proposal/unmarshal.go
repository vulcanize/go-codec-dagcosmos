package proposal

import (
	"io"
	"io/ioutil"

	"github.com/tendermint/tendermint/libs/protoio"

	"github.com/ipld/go-ipld-prime"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Decode provides an IPLD codec decode interface for Cosmos Proposal IPLDs.
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
	cp := new(tmproto.CanonicalProposal)
	if err := protoio.UnmarshalDelimited(src, cp); err != nil {
		return err
	}
	return DecodeCanonicalProposal(na, *cp)
}

// DecodeCanonicalProposal is like Decode, but it uses an input tendermint CanonicalProposal type
func DecodeCanonicalProposal(na ipld.NodeAssembler, cs tmproto.CanonicalProposal) error {
	ma, err := na.BeginMap(4)
	if err != nil {
		return err
	}
	for _, upFunc := range requiredUnpackFuncs {
		if err := upFunc(ma, cs); err != nil {
			return err
		}
	}
	return ma.Finish()
}

var requiredUnpackFuncs = []func(ipld.MapAssembler, tmproto.CanonicalProposal) error{
	unpackType,
	unpackHeight,
	unpackRound,
	unpackPOLRound,
	unpackBlockID,
	unpackTimestamp,
	unpackChainID,
}

func unpackType(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("SMType"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(int64(cp.Type))
}

func unpackHeight(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(cp.Height)
}

func unpackRound(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("Round"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(cp.Round)
}

func unpackPOLRound(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("POLRound"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(cp.POLRound)
}

func unpackBlockID(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("BlockID"); err != nil {
		return err
	}
	bidMA, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	blockID := types.BlockID{
		Hash: cp.BlockID.Hash,
		PartSetHeader: types.PartSetHeader{
			Total: cp.BlockID.PartSetHeader.Total,
			Hash:  cp.BlockID.PartSetHeader.Hash,
		},
	}
	return shared.UnpackBlockID(bidMA, blockID)
}

func unpackTimestamp(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("Timestamp"); err != nil {
		return err
	}
	tma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	return shared.UnpackTime(tma, cp.Timestamp)
}

func unpackChainID(ma ipld.MapAssembler, cp tmproto.CanonicalProposal) error {
	if err := ma.AssembleKey().AssignString("ChainID"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignString(cp.ChainID)
}
