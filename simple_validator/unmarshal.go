package params

import (
	"io"
	"io/ioutil"

	"github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/ipld/go-ipld-prime"
)

// Decode provides an IPLD codec decode interface for Cosmos SimpleValidator IPLDs.
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
	sv := new(types.SimpleValidator)
	if err := sv.Unmarshal(src); err != nil {
		return err
	}
	return DecodeSimpleValidator(na, *sv)
}

// DecodeSimpleValidator is like Decode, but it uses an input tendermint SimpleValidator type
func DecodeSimpleValidator(na ipld.NodeAssembler, sv types.SimpleValidator) error {
	ma, err := na.BeginMap(15)
	if err != nil {
		return err
	}
	for _, upFunc := range requiredUnpackFuncs {
		if err := upFunc(ma, sv); err != nil {
			return err
		}
	}
	return ma.Finish()
}

var requiredUnpackFuncs = []func(ipld.MapAssembler, types.SimpleValidator) error{
	unpackPubKey,
	unpackVotingPower,
}

func unpackPubKey(ma ipld.MapAssembler, sv types.SimpleValidator) error {
	if err := ma.AssembleKey().AssignString("PubKey"); err != nil {
		return err
	}
	pkBytes, err := sv.PubKey.Marshal()
	if err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(pkBytes)
}

func unpackVotingPower(ma ipld.MapAssembler, sv types.SimpleValidator) error {
	if err := ma.AssembleKey().AssignString("VotingPower"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(sv.VotingPower)
}
