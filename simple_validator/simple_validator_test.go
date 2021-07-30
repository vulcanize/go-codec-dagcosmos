package validator_test

import (
	"bytes"
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/crypto/ed25519"
	ce "github.com/tendermint/tendermint/crypto/encoding"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/simple_validator"
)

var (
	privKey             = ed25519.GenPrivKey()
	pubKey              = privKey.PubKey()
	pubKeyProto, _      = ce.PubKeyToProto(pubKey)
	testSimpleValidator = types.SimpleValidator{
		PubKey:      &pubKeyProto,
		VotingPower: 1337,
	}
	testSimpleValidatorEncoding, _ = testSimpleValidator.Marshal()
	simpleValidatorNode            ipld.Node
)

/* IPLD Schema
# This is what is actually included in the merkle tree
type SimpleValidator struct {
	PubKey      PubKey
	VotingPower Int
}

# Public key
type PubKey bytes
*/

func TestSimpleValidatorCodec(t *testing.T) {
	testSimpleValidatorDecode(t)
	testSimpleValidatorNodeContents(t)
	testSimpleValidatorEncode(t)
}

func testSimpleValidatorDecode(t *testing.T) {
	svBuilder := dagcosmos.Type.SimpleValidator.NewBuilder()
	svReader := bytes.NewReader(testSimpleValidatorEncoding)
	if err := validator.Decode(svBuilder, svReader); err != nil {
		t.Fatalf("unable to decode simple validator into an IPLD node: %v", err)
	}
	simpleValidatorNode = svBuilder.Build()
}

func testSimpleValidatorNodeContents(t *testing.T) {
	pkNode, err := simpleValidatorNode.LookupByString("PubKey")
	if err != nil {
		t.Fatalf("simple validator is missing PubKey: %v", err)
	}
	pkBytes, err := pkNode.AsBytes()
	if err != nil {
		t.Fatalf("simple validator PubKey should be of type Bytes: %v", err)
	}
	tmpk := new(crypto.PublicKey)
	if err := tmpk.Unmarshal(pkBytes); err != nil {
		t.Fatalf("simple validator unable to unmarshall PubKey into proto type: %v", err)
	}
	pk, err := ce.PubKeyFromProto(*tmpk)
	if err != nil {
		t.Fatalf("simple validator PubKey proto type cannot be converted to tendermint type: %v", err)
	}
	if !pk.Equals(pubKey) {
		t.Errorf("simple validator PubKey (%+v) does not match expected PubKey (%+v)", tmpk, pubKeyProto)
	}

	vpNode, err := simpleValidatorNode.LookupByString("VotingPower")
	if err != nil {
		t.Fatalf("simple validator is missing VotingPower: %v", err)
	}
	vp, err := vpNode.AsInt()
	if err != nil {
		t.Fatalf("simple validator VotingPower should be of type Int: %v", err)
	}
	if vp != testSimpleValidator.VotingPower {
		t.Errorf("simple validator VotingPower (%d) does not match expected VotingPower (%d)", vp, testSimpleValidator.VotingPower)
	}
}

func testSimpleValidatorEncode(t *testing.T) {
	svWriter := new(bytes.Buffer)
	if err := validator.Encode(simpleValidatorNode, svWriter); err != nil {
		t.Fatalf("unable to encode simple validator into writer: %v", err)
	}
	encodedSimpleValidator := svWriter.Bytes()
	if !bytes.Equal(encodedSimpleValidator, testSimpleValidatorEncoding) {
		t.Errorf("simple validator encoding (%x) does not match the expected encoding (%x)", encodedSimpleValidator, testSimpleValidatorEncoding)
	}
}
