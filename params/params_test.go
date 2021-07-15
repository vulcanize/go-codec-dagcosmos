package params_test

import (
	"bytes"
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/params"
)

var (
	testParams = types.HashedParams{
		BlockMaxGas:   1337,
		BlockMaxBytes: 1338,
	}
	hashedParamsEncoding, _ = testParams.Marshal()
	paramsNode              ipld.Node
)

/* IPLD Schema
# HashedParams is a subset of ConsensusParams that is included in the consensus encoding
# It is hashed into the Header.ConsensusHash.
type HashedParams struct {
	BlockMaxBytes Int
	BlockMaxGas   Int
}
*/

func TestHashedParamsCodec(t *testing.T) {
	testHashedParamsDecode(t)
	testHashedParamsNodeContents(t)
	testHashedParamsEncode(t)
}

func testHashedParamsDecode(t *testing.T) {
	hashedParamsBuilder := dagcosmos.Type.HashedParams.NewBuilder()
	hashedParamsReader := bytes.NewReader(hashedParamsEncoding)
	if err := params.Decode(hashedParamsBuilder, hashedParamsReader); err != nil {
		t.Fatalf("unable to decode hashed params into an IPLD node: %v", err)
	}
	paramsNode = hashedParamsBuilder.Build()
}

func testHashedParamsNodeContents(t *testing.T) {
	bmbNode, err := paramsNode.LookupByString("BlockMaxBytes")
	if err != nil {
		t.Fatalf("hashed params is missing BlockMaxBytes: %v", err)
	}
	bmb, err := bmbNode.AsInt()
	if err != nil {
		t.Fatalf("hashed params BlockMaxBytes should be of type Int: %v", err)
	}
	if bmb != testParams.BlockMaxBytes {
		t.Errorf("hashed params BlockMaxBytes (%d) does not match expected BlockMaxBytes (%d)", bmb, testParams.BlockMaxBytes)
	}

	bmgNode, err := paramsNode.LookupByString("BlockMaxGas")
	if err != nil {
		t.Fatalf("hashed params is missing BlockMaxGas: %v", err)
	}
	bmg, err := bmgNode.AsInt()
	if err != nil {
		t.Fatalf("hashed params BlockMaxGas should be of type Int: %v", err)
	}
	if bmg != testParams.BlockMaxGas {
		t.Errorf("hashed params BlockMaxGas (%d) does not match expected BlockMaxGas (%d)", bmg, testParams.BlockMaxGas)
	}
}

func testHashedParamsEncode(t *testing.T) {
	hashedParamsWriter := new(bytes.Buffer)
	if err := params.Encode(paramsNode, hashedParamsWriter); err != nil {
		t.Fatalf("unable to encode hashed params into writer: %v", err)
	}
	encodedHashedParamsBytes := hashedParamsWriter.Bytes()
	if !bytes.Equal(encodedHashedParamsBytes, hashedParamsEncoding) {
		t.Errorf("hashed params encoding (%x) does not match the expected encoding (%x)", encodedHashedParamsBytes, hashedParamsEncoding)
	}
}
