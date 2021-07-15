package commit_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/commit"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	valAddr       = shared.RandomAddr()
	sig           = []byte("mockSignatureBytes")
	timestamp     = time.Now()
	blockIDFlag   = types.BlockIDFlag(2)
	testCommitSig = types.CommitSig{
		ValidatorAddress: valAddr,
		Signature:        sig,
		Timestamp:        timestamp,
		BlockIDFlag:      blockIDFlag,
	}
	commitSigProto       = testCommitSig.ToProto()
	commitSigEncoding, _ = commitSigProto.Marshal()
	commitNode           ipld.Node
)

/* IPLD Schema
# CommitSig is a part of the Vote included in a Commit.
# These are the leaf values in the merkle tree referenced by LastCommitHash
type CommitSig struct {
	BlockIDFlag      BlockIDFlag
	ValidatorAddress Address
	Timestamp        Time
	Signature        Signature
}

# BlockIDFlag is a single byte flag
type BlockIDFlag enum {
  | BlockIDFlagUnknown ("0")
  | BlockIDFlagAbsent ("1")
  | BlockIDFlagCommit ("2")
  | BlockIDFlagNil ("3")
} representation int

type Signature bytes
*/

func TestCommitSigCodec(t *testing.T) {
	testCommitSigDecode(t)
	testCommitSigNodeContents(t)
	testCommitSigEncode(t)
}

func testCommitSigDecode(t *testing.T) {
	commitSigBuilder := dagcosmos.Type.CommitSig.NewBuilder()
	commitSigReader := bytes.NewReader(commitSigEncoding)
	if err := commit.Decode(commitSigBuilder, commitSigReader); err != nil {
		t.Fatalf("unable to decode commit sig into an IPLD node: %v", err)
	}
	commitNode = commitSigBuilder.Build()
}

func testCommitSigNodeContents(t *testing.T) {
	blockIDFlagNode, err := commitNode.LookupByString("BlockIDFlag")
	if err != nil {
		t.Fatalf("commit sig is missing BlockIDFlag: %v", err)
	}
	blockIDFlagInt, err := blockIDFlagNode.AsInt()
	if err != nil {
		t.Fatalf("commit sig BlockIDFlag should be of type Int: %v", err)
	}
	if types.BlockIDFlag(blockIDFlagInt) != blockIDFlag {
		t.Errorf("commit sig BlockIDFlag (%d) does not match expected BlockIDFlag (%d)", blockIDFlagInt, blockIDFlag)
	}

	vaNode, err := commitNode.LookupByString("ValidatorAddress")
	if err != nil {
		t.Fatalf("commit sig is missing ValidatorAddress: %v", err)
	}
	vaBytes, err := vaNode.AsBytes()
	if err != nil {
		t.Fatalf("commit sig ValidatorAddress should be of type Bytes: %v", err)
	}
	if !bytes.Equal(vaBytes, valAddr) {
		t.Errorf("commit sig ValidatorAddress (%x) does not match expected ValidatorAddress (%x)", vaBytes, valAddr)
	}

	timeNode, err := commitNode.LookupByString("Timestamp")
	if err != nil {
		t.Fatalf("commit sig is missing Timestamp: %v", err)
	}
	ti, err := shared.PackTime(timeNode)
	if err != nil {
		t.Fatalf("commit sig Timestamp could not be unpacked into time.Time: %v", err)
	}
	if !ti.Equal(timestamp) {
		t.Errorf("commit sig Timestamp (%s) does not match expected Timestamp (%s)", ti.String(), timestamp.String())
	}

	sigNode, err := commitNode.LookupByString("Signature")
	if err != nil {
		t.Fatalf("commit sig is missing Signature: %v", err)
	}
	sigBytes, err := sigNode.AsBytes()
	if err != nil {
		t.Fatalf("commit sig Signature should be of type Bytes: %v", err)
	}
	if !bytes.Equal(sigBytes, sig) {
		t.Errorf("commit sig Signature (%x) does not match expected Signature (%x)", sigBytes, sig)
	}
}

func testCommitSigEncode(t *testing.T) {
	commitSigWriter := new(bytes.Buffer)
	if err := commit.Encode(commitNode, commitSigWriter); err != nil {
		t.Fatalf("unable to encode commit sig into writer: %v", err)
	}
	encodedCommitSigBytes := commitSigWriter.Bytes()
	if !bytes.Equal(encodedCommitSigBytes, commitSigEncoding) {
		t.Errorf("commit sig encoding (%x) does not match the expected encoding (%x)", encodedCommitSigBytes, commitSigEncoding)
	}
}
