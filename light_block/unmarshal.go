package light_block

import (
	"io"
	"io/ioutil"

	"github.com/ipld/go-ipld-prime"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/vulcanize/go-codec-dagcosmos/header"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
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
	unpackSignedHeader,
	unpackValidatorSet,
}

func unpackSignedHeader(ma ipld.MapAssembler, lb types.LightBlock) error {
	if err := ma.AssembleKey().AssignString("SignedHeader"); err != nil {
		return err
	}
	shMA, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := shMA.AssembleKey().AssignString("Header"); err != nil {
		return err
	}
	headerNode := shMA.AssembleValue().Prototype().NewBuilder()
	if err := header.DecodeHeader(headerNode, *lb.Header); err != nil {
		return err
	}
	if err := shMA.AssembleValue().AssignNode(headerNode.Build()); err != nil {
		return err
	}
	if err := shMA.AssembleKey().AssignString("Commit"); err != nil {
		return err
	}
	commitMA, err := shMA.AssembleValue().BeginMap(4)
	if err != nil {
		return err
	}
	if err := shared.UnpackCommit(commitMA, *lb.Commit); err != nil {
		return err
	}
	return shMA.Finish()
}

func unpackValidatorSet(ma ipld.MapAssembler, lb types.LightBlock) error {
	if err := ma.AssembleKey().AssignString("ValidatorSet"); err != nil {
		return err
	}
	valSetMA, err := ma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := valSetMA.AssembleKey().AssignString("Validators"); err != nil {
		return err
	}
	valsLA, err := valSetMA.AssembleValue().BeginList(int64(len(lb.ValidatorSet.Validators)))
	if err != nil {
		return err
	}
	for _, validator := range lb.ValidatorSet.Validators {
		valMA, err := valsLA.AssembleValue().BeginMap(4)
		if err != nil {
			return err
		}
		if err := shared.UnpackValidator(valMA, *validator); err != nil {
			return err
		}
	}
	if err := valsLA.Finish(); err != nil {
		return err
	}
	return valSetMA.Finish()
}
