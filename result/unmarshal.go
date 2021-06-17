package params

import (
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/ipld/go-ipld-prime"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Decode provides an IPLD codec decode interface for Tendermint Result IPLDs.
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
	res := new(abci.ResponseDeliverTx)
	if err := res.Unmarshal(src); err != nil {
		return err
	}
	return DecodeParams(na, *res)
}

// DecodeParams is like Decode, but it uses an input tendermint HashedParams type
func DecodeParams(na ipld.NodeAssembler, hp abci.ResponseDeliverTx) error {
	ma, err := na.BeginMap(4)
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

var requiredUnpackFuncs = []func(ipld.MapAssembler, abci.ResponseDeliverTx) error{
	unpackCode,
	unpackData,
	unpackGasWanted,
	unpackGasUsed,
}

func unpackCode(ma ipld.MapAssembler, res abci.ResponseDeliverTx) error {
	if err := ma.AssembleKey().AssignString("Code"); err != nil {
		return err
	}
	codeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(codeBytes, res.Code)
	return ma.AssembleValue().AssignBytes(codeBytes)
}

func unpackData(ma ipld.MapAssembler, res abci.ResponseDeliverTx) error {
	if err := ma.AssembleKey().AssignString("Data"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignBytes(res.Data)
}

func unpackGasWanted(ma ipld.MapAssembler, res abci.ResponseDeliverTx) error {
	if err := ma.AssembleKey().AssignString("GasWanted"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(res.GasWanted)
}

func unpackGasUsed(ma ipld.MapAssembler, res abci.ResponseDeliverTx) error {
	if err := ma.AssembleKey().AssignString("GasUsed"); err != nil {
		return err
	}
	return ma.AssembleValue().AssignInt(res.GasUsed)
}
