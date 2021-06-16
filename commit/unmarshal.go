package commit

import (
	"io"
	"io/ioutil"

	"github.com/ipld/go-ipld-prime"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// Decode provides an IPLD codec decode interface for Cosmos CommitSig IPLDs.
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
	var pcs tmproto.CommitSig

	if err := pcs.Unmarshal(src); err != nil {
		return err
	}
	cs := new(types.CommitSig)
	if err := cs.FromProto(pcs); err != nil {
		return err
	}
	return DecodeCommitSig(na, *cs)
}

// DecodeCommitSig is like Decode, but it uses an input tendermint CommitSig type
func DecodeCommitSig(na ipld.NodeAssembler, cs types.CommitSig) error {
	ma, err := na.BeginMap(15)
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

var requiredUnpackFuncs = []func(ipld.MapAssembler, types.CommitSig) error{
	unpackBlockIDFlag,
	unpackValidatorAddress,
	unpackTimestamp,
	unpackSignature,
}

func unpackBlockIDFlag(ma ipld.MapAssembler, cs types.CommitSig) error {
	if err := ma.AssembleKey().AssignString("BlockIDFlag"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(int64(cs.BlockIDFlag))
}

func unpackValidatorAddress(ma ipld.MapAssembler, cs types.CommitSig) error {
	if err := ma.AssembleKey().AssignString("ValidatorAddress"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(cs.ValidatorAddress)
}

func unpackTimestamp(ma ipld.MapAssembler, cs types.CommitSig) error {
	if err := ma.AssembleKey().AssignString("Timestamp"); err != nil {
		return err
	}
	tma, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	return shared.UnpackTime(tma, cs.Timestamp)
}

func unpackSignature(ma ipld.MapAssembler, cs types.CommitSig) error {
	if err := ma.AssembleKey().AssignString("Signature"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(cs.Signature)
}
