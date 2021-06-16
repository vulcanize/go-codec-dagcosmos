package evidence

import (
	"io"
	"io/ioutil"

	"github.com/ipld/go-ipld-prime"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/vulcanize/go-codec-dagcosmos/light_block"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Decode provides an IPLD codec decode interface for Cosmos Evidence IPLDs.
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
	tmle := new(tmproto.LightClientAttackEvidence)
	if err := tmle.Unmarshal(src); err == nil {
		le, err := types.LightClientAttackEvidenceFromProto(tmle)
		if err != nil {
			return err
		}
		return DecodeLightEvidence(na, *le)
	}
	tmdve := new(tmproto.DuplicateVoteEvidence)
	if err := tmdve.Unmarshal(src); err != nil {
		return err
	}
	de, err := types.DuplicateVoteEvidenceFromProto(tmdve)
	if err != nil {
		return err
	}
	return DecodeDuplicateEvidence(na, *de)
}

// DecodeLightEvidence is like Decode, but it uses an input tendermint LightClientAttackEvidence type
func DecodeLightEvidence(na ipld.NodeAssembler, e types.LightClientAttackEvidence) error {
	ma, err := na.BeginMap(15)
	if err != nil {
		return err
	}
	for _, upFunc := range leRequiredUnpackFuncs {
		if err := upFunc(ma, e); err != nil {
			return err
		}
	}
	return ma.Finish()
}

// DecodeDuplicateEvidence is like Decode, but it uses an input tendermint DuplicateVoteEvidence type
func DecodeDuplicateEvidence(na ipld.NodeAssembler, e types.DuplicateVoteEvidence) error {
	ma, err := na.BeginMap(15)
	if err != nil {
		return err
	}
	for _, upFunc := range deRequiredUnpackFuncs {
		if err := upFunc(ma, e); err != nil {
			return err
		}
	}
	return ma.Finish()
}

var deRequiredUnpackFuncs = []func(ipld.MapAssembler, types.DuplicateVoteEvidence) error{
	unpackVoteA,
	unpackVoteB,
	unpackDETotalVotingPower,
	unpackValidatorPower,
	unpackDETimestamp,
}

var leRequiredUnpackFuncs = []func(ipld.MapAssembler, types.LightClientAttackEvidence) error{
	unpackConflictingBlock,
	unpackCommonHeight,
	unpackByzantineValidators,
	unpackLETotalVotingPower,
	unpackLETimestamp,
}

func unpackVoteA(ma ipld.MapAssembler, de types.DuplicateVoteEvidence) error {
	if err := ma.AssembleKey().AssignString("VoteA"); err != nil {
		return err
	}
	vama, err := ma.AssembleValue().BeginMap(8)
	if err != nil {
		return err
	}
	return shared.UnpackVote(vama, *de.VoteA)
}

func unpackVoteB(ma ipld.MapAssembler, de types.DuplicateVoteEvidence) error {
	if err := ma.AssembleKey().AssignString("VoteB"); err != nil {
		return err
	}
	vama, err := ma.AssembleValue().BeginMap(8)
	if err != nil {
		return err
	}
	return shared.UnpackVote(vama, *de.VoteB)
}

func unpackDETotalVotingPower(ma ipld.MapAssembler, de types.DuplicateVoteEvidence) error {
	if err := ma.AssembleKey().AssignString("TotalVotingPower"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(de.TotalVotingPower)
}

func unpackValidatorPower(ma ipld.MapAssembler, de types.DuplicateVoteEvidence) error {
	if err := ma.AssembleKey().AssignString("ValidatorPower"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(de.ValidatorPower)
}

func unpackDETimestamp(ma ipld.MapAssembler, de types.DuplicateVoteEvidence) error {
	if err := ma.AssembleKey().AssignString("Timestamp"); err != nil {
		return err
	}
	tma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	return shared.UnpackTime(tma, de.Timestamp)
}

func unpackConflictingBlock(ma ipld.MapAssembler, le types.LightClientAttackEvidence) error {
	if err := ma.AssembleKey().AssignString("ConflictingBlock"); err != nil {
		return err
	}
	return light_block.DecodeLightBlock(ma.AssembleValue(), *le.ConflictingBlock)
}

func unpackCommonHeight(ma ipld.MapAssembler, le types.LightClientAttackEvidence) error {
	if err := ma.AssembleKey().AssignString("CommonHeight"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(le.CommonHeight)
}

func unpackByzantineValidators(ma ipld.MapAssembler, le types.LightClientAttackEvidence) error {
	if err := ma.AssembleKey().AssignString("ByzantineValidators"); err != nil {
		return err
	}
	bvLA, err := ma.AssembleValue().BeginList(int64(len(le.ByzantineValidators)))
	if err != nil {
		return err
	}
	for _, bv := range le.ByzantineValidators {
		vaMA, err := bvLA.AssembleValue().BeginMap(4)
		if err != nil {
			return err
		}
		if err := shared.UnpackValidator(vaMA, *bv); err != nil {
			return err
		}
	}
	return bvLA.Finish()
}

func unpackLETotalVotingPower(ma ipld.MapAssembler, le types.LightClientAttackEvidence) error {
	if err := ma.AssembleKey().AssignString("TotalVotingPower"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(le.TotalVotingPower)
}

func unpackLETimestamp(ma ipld.MapAssembler, le types.LightClientAttackEvidence) error {
	if err := ma.AssembleKey().AssignString("Timestamp"); err != nil {
		return err
	}
	tma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	return shared.UnpackTime(tma, le.Timestamp)
}
