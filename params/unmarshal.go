package params

import (
	"io"
	"io/ioutil"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

// Decode provides an IPLD codec decode interface for Cosmos Params IPLDs.
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
	var hp types.HashedParams
	if err := hp.Unmarshal(src); err != nil {
		return err
	}
	return DecodeParams(na, hp)
}

// DecodeParams is like Decode, but it uses an input tendermint HashedParams type
func DecodeParams(na ipld.NodeAssembler, hp types.HashedParams) error {
	ma, err := na.BeginMap(15)
	if err != nil {
		return err
	}
	for _, upFunc := range requiredUnpackFuncs {
		if err := upFunc(ma, hp); err != nil {
			return err
		}
	}
	return ma.Finish()
}

var requiredUnpackFuncs = []func(ipld.MapAssembler, types.HashedParams) error{
	unpackMaxGas,
	unpackMaxBytes,
}

func unpackMaxBytes(ma ipld.MapAssembler, hp types.HashedParams) error {
	if err := ma.AssembleKey().AssignString("BlockMaxBytes"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(hp.BlockMaxBytes)
}

func unpackMaxGas(ma ipld.MapAssembler, hp types.HashedParams) error {
	if err := ma.AssembleKey().AssignString("BlockMaxGas"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(hp.BlockMaxGas)
}
