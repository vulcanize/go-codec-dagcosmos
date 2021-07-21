package result_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/ipld/go-ipld-prime"
	abci "github.com/tendermint/tendermint/abci/types"
	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/result"
)

var (
	testResult = abci.ResponseDeliverTx{
		Code:      1335,
		Data:      []byte("testData"),
		GasWanted: 1337,
		GasUsed:   1336,
	}
	testResultEncoding, _ = testResult.Marshal()
	resultNode            ipld.Node
)

/* IPLD Schema
// ResponseDeliverTx is an ABCI response to DeliverTx requests
// this includes the consensus fields only, this is included in the merkle tree referenced by LastResulHash in the Header
type ResponseDeliverTx struct {
	Code      Uint
	Data      Bytes
	GasWanted Int
	GasUsed   Int
}
*/

func TestResultCodec(t *testing.T) {
	testResultDecode(t)
	testResultNodeContents(t)
	testResultEncode(t)
}

func testResultDecode(t *testing.T) {
	resBuilder := dagcosmos.Type.ResponseDeliverTx.NewBuilder()
	resReader := bytes.NewReader(testResultEncoding)
	if err := result.Decode(resBuilder, resReader); err != nil {
		t.Fatalf("unable to decode result into an IPLD node: %v", err)
	}
	resultNode = resBuilder.Build()
}

func testResultNodeContents(t *testing.T) {
	codeNode, err := resultNode.LookupByString("Code")
	if err != nil {
		t.Fatalf("result is missing Code: %v", err)
	}
	codeBytes, err := codeNode.AsBytes()
	if err != nil {
		t.Fatalf("result Code should be of type Bytes: %v", err)
	}
	code := binary.BigEndian.Uint32(codeBytes)
	if code != testResult.Code {
		t.Errorf("result Code (%d) does not match expected Code (%d)", code, testResult.Code)
	}

	dataNode, err := resultNode.LookupByString("Data")
	if err != nil {
		t.Fatalf("result is missing Data: %v", err)
	}
	data, err := dataNode.AsBytes()
	if err != nil {
		t.Fatalf("result Data should be of type Bytes: %v", err)
	}
	if !bytes.Equal(data, testResult.Data) {
		t.Errorf("result Data (%x) does not match expected Data (%x)", data, testResult.Data)
	}

	gwNode, err := resultNode.LookupByString("GasWanted")
	if err != nil {
		t.Fatalf("result is missing GasWanted: %v", err)
	}
	gw, err := gwNode.AsInt()
	if err != nil {
		t.Fatalf("result GasWanted should be of type Int: %v", err)
	}
	if gw != testResult.GasWanted {
		t.Errorf("result GasWanted (%d) does not match expected GasWanted (%d)", gw, testResult.GasWanted)
	}

	guNode, err := resultNode.LookupByString("GasUsed")
	if err != nil {
		t.Fatalf("result is missing GasUsed: %v", err)
	}
	gu, err := guNode.AsInt()
	if err != nil {
		t.Fatalf("result GasUsed should be of type Int: %v", err)
	}
	if gu != testResult.GasUsed {
		t.Errorf("result GasUsed (%d) does not match expected GasUsed (%d)", gw, testResult.GasUsed)
	}
}

func testResultEncode(t *testing.T) {
	resWriter := new(bytes.Buffer)
	if err := result.Encode(resultNode, resWriter); err != nil {
		t.Fatalf("unable to encode result into writer: %v", err)
	}
	encodedResult := resWriter.Bytes()
	if !bytes.Equal(encodedResult, testResultEncoding) {
		t.Errorf("result encoding (%x) does not match the expected encoding (%x)", encodedResult, testResultEncoding)
	}
}
