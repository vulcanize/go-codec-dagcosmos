package light_block_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	commit2 "github.com/vulcanize/go-codec-dagcosmos/commit"
	"github.com/vulcanize/go-codec-dagcosmos/header"
	"github.com/vulcanize/go-codec-dagcosmos/light_block"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	lightBlock = types.LightBlock{
		SignedHeader: signedHeader,
		ValidatorSet: validatorSet,
	}
	signedHeader = &types.SignedHeader{
		Header: testHeader,
		Commit: testCommit,
	}
	testHeader = &types.Header{
		Version: tmversion.Consensus{
			Block: 11,
			App:   2,
		},
		ChainID: "mockChainID",
		Height:  1337,
		Time:    time.Now(),
		LastBlockID: types.BlockID{
			Hash: shared.RandomHash(),
			PartSetHeader: types.PartSetHeader{
				Hash:  shared.RandomHash(),
				Total: 1338,
			},
		},
		LastCommitHash:     shared.RandomHash(),
		DataHash:           shared.RandomHash(),
		ValidatorsHash:     validatorHash,
		NextValidatorsHash: shared.RandomHash(),
		ConsensusHash:      shared.RandomHash(),
		AppHash:            shared.RandomHash(),
		LastResultsHash:    shared.RandomHash(),
		EvidenceHash:       shared.RandomHash(),
		ProposerAddress:    shared.RandomAddr(),
	}
	testHeaderProto       = testHeader.ToProto()
	testHeaderEncoding, _ = testHeaderProto.Marshal()
	privKey1              = ed25519.GenPrivKey()
	pubKey1               = privKey1.PubKey()
	privKey2              = ed25519.GenPrivKey()
	pubKey2               = privKey2.PubKey()
	privKey3              = ed25519.GenPrivKey()
	pubKey3               = privKey3.PubKey()
	validator1            = &types.Validator{
		Address:          shared.RandomAddr(),
		PubKey:           pubKey1,
		VotingPower:      13337,
		ProposerPriority: 13338,
	}
	validator1Proto, _    = validator1.ToProto()
	validator1Encoding, _ = validator1Proto.Marshal()
	validator2            = &types.Validator{
		Address:          shared.RandomAddr(),
		PubKey:           pubKey2,
		VotingPower:      133337,
		ProposerPriority: 133338,
	}
	validator2Proto, _    = validator2.ToProto()
	validator2Encoding, _ = validator2Proto.Marshal()
	validators            = []*types.Validator{
		validator1,
		validator2,
	}
	validatorsEncoding = [][]byte{
		validator1Encoding,
		validator2Encoding,
	}
	testProposer = &types.Validator{
		Address:          shared.RandomAddr(),
		PubKey:           pubKey3,
		VotingPower:      1333337,
		ProposerPriority: 1333338,
	}
	testProposerProto, _    = testProposer.ToProto()
	testProposerEncoding, _ = testProposerProto.Marshal()
	validatorSet            = &types.ValidatorSet{
		Validators: validators,
		Proposer:   testProposer,
	}
	validatorHash = validatorSet.Hash()
	blockHash     = testHeader.Hash()
	testCommit    = &types.Commit{
		Height: 1337,
		Round:  4,
		BlockID: types.BlockID{
			Hash: blockHash,
			PartSetHeader: types.PartSetHeader{
				Hash:  shared.RandomHash(),
				Total: 13333337,
			},
		},
		Signatures: signatures,
	}
	testCommitProto       = testCommit.ToProto()
	testCommitEncoding, _ = testCommitProto.Marshal()
	signatures            = []types.CommitSig{
		types.CommitSig{
			BlockIDFlag:      types.BlockIDFlagCommit,
			ValidatorAddress: shared.RandomAddr(),
			Timestamp:        time.Time{},
			Signature:        shared.RandomSig(),
		},
		types.CommitSig{
			BlockIDFlag:      types.BlockIDFlagCommit,
			ValidatorAddress: shared.RandomAddr(),
			Timestamp:        time.Time{},
			Signature:        shared.RandomSig(),
		},
	}
	lightBlockProto, _    = lightBlock.ToProto()
	lightBlockEncoding, _ = lightBlockProto.Marshal()
	lightBlockNode        ipld.Node
)

/* IPLD Schema
# LightBlock is a SignedHeader and a ValidatorSet.
# It is the basis of the light client
type LightBlock struct {
	SignedHeader SignedHeader
	ValidatorSet  ValidatorSet
}

# SignedHeader is a header along with the commits that prove it.
type SignedHeader struct {
	Header Header
	Commit Commit
}

# Volatile state for each Validator
# NOTE: The Address and ProposerPriority is not included in Validator.Hash();
# make sure to update that method if changes are made here
type Validator struct {
	Address     Address # this should be remove since it isn't included in the content hahs?
	PubKey      PubKey
	VotingPower Int
	ProposerPriority Int # this should be removed since it isn't included in the content hash?
}

# ValidatorSet represent a set of Validators at a given height.
#
# The validators can be fetched by address or index.
# The index is in order of .VotingPower, so the indices are fixed for all
# rounds of a given blockchain height - ie. the validators are sorted by their
# voting power (descending). Secondary index - .Address (ascending).
type ValidatorSet struct {
	Validators []Validator
	Proposer   Validator
}
*/

func TestLightBlockCodec(t *testing.T) {
	testLightBlockDecode(t)
	testLightBlockNodeContents(t)
	testLightBlockEncode(t)
}

func testLightBlockDecode(t *testing.T) {
	lbBuilder := dagcosmos.Type.LightBlock.NewBuilder()
	lbReader := bytes.NewReader(lightBlockEncoding)
	if err := light_block.Decode(lbBuilder, lbReader); err != nil {
		t.Fatalf("unable to decode light block into an IPLD node: %v", err)
	}
	lightBlockNode = lbBuilder.Build()
}

func testLightBlockNodeContents(t *testing.T) {
	shNode, err := lightBlockNode.LookupByString("SignedHeader")
	if err != nil {
		t.Fatalf("light block is missing SignedHeader: %v", err)
	}
	hNode, err := shNode.LookupByString("Header")
	if err != nil {
		t.Fatalf("light block SignedHeader is missing Header: %v", err)
	}
	h := new(types.Header)
	if err := header.EncodeHeader(h, hNode); err != nil {
		t.Fatalf("light block SignedHeader Header cannot be packed: %v", err)
	}
	hProto := h.ToProto()
	encodedHeader, err := hProto.Marshal()
	if err != nil {
		t.Fatalf("light block SignedHeader Header proto type cannot be marshalled: %v", err)
	}
	if !bytes.Equal(encodedHeader, testHeaderEncoding) {
		t.Errorf("light block SignedHeader Header (%+v) does not match expected Header (%+v)", h, testHeader)
	}
	commitNode, err := shNode.LookupByString("Commit")
	if err != nil {
		t.Fatalf("light block SignedHeader is missing Commit: %v", err)
	}
	commit, err := commit2.PackCommit(commitNode)
	if err != nil {
		t.Fatalf("light block SignedHeader Commit cannot be packed: %v", err)
	}
	commitProto := commit.ToProto()
	commitEncoding, err := commitProto.Marshal()
	if err != nil {
		t.Fatalf("light block SignedHeader Commit proto type cannot be marshalled: %v", err)
	}
	if !bytes.Equal(commitEncoding, testCommitEncoding) {
		t.Errorf("light block SignedHeader Commit (%+v) does not match expected Commit (%+v)", commit, testCommit)
	}

	vSetNode, err := lightBlockNode.LookupByString("ValidatorSet")
	if err != nil {
		t.Fatalf("light block is missing ValidatorSet: %v", err)
	}
	validatorsNode, err := vSetNode.LookupByString("Validators")
	if err != nil {
		t.Fatalf("light block ValidatorSet is missing Validators: %v", err)
	}
	valsIT := validatorsNode.ListIterator()
	for !valsIT.Done() {
		i, validatorNode, err := valsIT.Next()
		if err != nil {
			t.Fatalf("light block ValidatorSet Validators iterator error: %v", err)
		}
		validator, err := shared.PackValidator(validatorNode)
		if err != nil {
			t.Fatalf("light block ValidatorSet Validators Validator cannot be packed: %v", err)
		}
		validatorProto, err := validator.ToProto()
		if err != nil {
			t.Fatalf("light block ValidatorSet Validators Validator proto type cannot be converted to proto type: %v", err)
		}
		encodedValidator, err := validatorProto.Marshal()
		if err != nil {
			t.Fatalf("light block ValidatorSet Validators Validator proto type cannot be marshalled: %v", err)
		}
		if !bytes.Equal(encodedValidator, validatorsEncoding[i]) {
			t.Errorf("light block ValidatorSet Validators Validator (%+v) does not match expected Validator (%x)", validator, validators[i])
		}
	}

	proposerNode, err := vSetNode.LookupByString("Proposer")
	if err != nil {
		t.Fatalf("light block ValidatorSet is missing Proposer: %v", err)
	}
	proposer, err := shared.PackValidator(proposerNode)
	if err != nil {
		t.Fatalf("light block ValidatorSet Proposer cannot be packed: %v", err)
	}
	proposerProto, err := proposer.ToProto()
	if err != nil {
		t.Fatalf("light block ValidatorSet Proposer cannot be converted to proto type: %v", err)
	}
	proposerEncoding, err := proposerProto.Marshal()
	if err != nil {
		t.Fatalf("light block ValidatorSet Proposer proto type cannot be marshalled: %v", err)
	}
	if !bytes.Equal(proposerEncoding, testProposerEncoding) {
		t.Errorf("light block ValidatorSet Proposer (%+v) does not match expected Proposer (%+v)", proposer, testProposer)
	}
}

func testLightBlockEncode(t *testing.T) {
	lbWriter := new(bytes.Buffer)
	if err := light_block.Encode(lightBlockNode, lbWriter); err != nil {
		t.Fatalf("unable to encode light client attack evidence into writer: %v", err)
	}
	encodedLightBlock := lbWriter.Bytes()
	if !bytes.Equal(encodedLightBlock, lightBlockEncoding) {
		t.Errorf("light block encoding (%x) does not match the expected encoding (%x)", encodedLightBlock, lightBlockEncoding)
	}
}
