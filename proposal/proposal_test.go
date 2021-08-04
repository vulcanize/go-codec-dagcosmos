package proposal_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/libs/protoio"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/proposal"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	testProposal = &types.CanonicalProposal{
		Type:      1,
		Height:    1337,
		Round:     5,
		POLRound:  4,
		BlockID:   canonicalBlockID,
		Timestamp: time.Now(),
		ChainID:   "test_chain_id",
	}
	canonicalBlockID = &types.CanonicalBlockID{
		Hash: shared.RandomHash(),
		PartSetHeader: types.CanonicalPartSetHeader{
			Hash:  shared.RandomHash(),
			Total: 1338,
		},
	}
	testProposalEncoding, _ = protoio.MarshalDelimited(testProposal)
	canonicalBlockIDEnc, _  = canonicalBlockID.Marshal()
	proposalNode            ipld.Node
)

/* IPLD Schema
# Proposal defines a block proposal for the consensus.
type Proposal struct {
	SMType    SignedMsgType
	Height    Int
	Round     Int # there can not be greater than 2_147_483_647 rounds
	POLRound  Int # -1 if null.
	BlockID   BlockID
	Timestamp Time
	ChainID   String
}

# SignedMsgType is the type of signed message in the consensus.
type SignedMsgType enum {
	| UnknownType ("0")
	| PrevoteType ("1")
	| PrecommitType ("2")
	| ProposalType ("32")
} representation int

# BlockID contains two distinct Merkle roots of the block.
# The BlockID includes these two hashes, as well as the number of parts (ie. len(MakeParts(block)))
type BlockID struct {
	Hash          HeaderCID
	PartSetHeader PartSetHeader
}

# PartSetHeader is used for secure gossiping of the block during consensus
# It contains the Merkle root of the complete serialized block cut into parts (ie. MerkleRoot(MakeParts(block))).
type PartSetHeader struct {
	Total Uint
	Hash  PartTreeCID
}
*/

func TestProposalCodec(t *testing.T) {
	testProposalDecode(t)
	testProposalNodeContents(t)
	testProposalEncode(t)
}

func testProposalDecode(t *testing.T) {
	proposalBuilder := dagcosmos.Type.Proposal.NewBuilder()
	proposalReader := bytes.NewReader(testProposalEncoding)
	if err := proposal.Decode(proposalBuilder, proposalReader); err != nil {
		t.Fatalf("unable to decode proposal into an IPLD node: %v", err)
	}
	proposalNode = proposalBuilder.Build()
}

/*
type Proposal struct {
	SMType    SignedMsgType
	Height    Int
	Round     Int # there can not be greater than 2_147_483_647 rounds
	POLRound  Int # -1 if null.
	BlockID   BlockID
	Timestamp Time
	ChainID   String
}
*/

func testProposalNodeContents(t *testing.T) {
	smTypeNode, err := proposalNode.LookupByString("SMType")
	if err != nil {
		t.Fatalf("proposal is missing PubKey: %v", err)
	}
	smType, err := smTypeNode.AsInt()
	if err != nil {
		t.Fatalf("proposal SMType should be of type Int: %v", err)
	}
	if types.SignedMsgType(smType) != testProposal.Type {
		t.Errorf("proposal SMType (%d) does not match expected SMType (%d)", smType, testProposal.Type)
	}

	heightNode, err := proposalNode.LookupByString("Height")
	if err != nil {
		t.Fatalf("proposal is missing Height: %v", err)
	}
	height, err := heightNode.AsInt()
	if err != nil {
		t.Fatalf("proposal Height should be of type Int: %v", err)
	}
	if height != testProposal.Height {
		t.Errorf("proposal Height (%d) does not match expected Height (%d)", height, testProposal.Height)
	}

	roundNode, err := proposalNode.LookupByString("Round")
	if err != nil {
		t.Fatalf("proposal is missing Round: %v", err)
	}
	round, err := roundNode.AsInt()
	if err != nil {
		t.Fatalf("proposal Round should be of type Int: %v", err)
	}
	if round != testProposal.Round {
		t.Errorf("proposal Round (%d) does not match expected Round (%d)", round, testProposal.Round)
	}

	polRoundNode, err := proposalNode.LookupByString("POLRound")
	if err != nil {
		t.Fatalf("proposal is missing POLRound: %v", err)
	}
	polRound, err := polRoundNode.AsInt()
	if err != nil {
		t.Fatalf("proposal POLRound should be of type Int: %v", err)
	}
	if polRound != testProposal.POLRound {
		t.Errorf("proposal POLRound (%d) does not match expected POLRound (%d)", polRound, testProposal.POLRound)
	}

	blockIDNode, err := proposalNode.LookupByString("BlockID")
	if err != nil {
		t.Fatalf("proposal is missing BlockID: %v", err)
	}
	blockID, err := shared.PackBlockID(blockIDNode)
	if err != nil {
		t.Fatalf("proposal BlockID cannot be packed: %v", err)
	}
	gotCanonicalBlockID := &types.CanonicalBlockID{
		Hash: blockID.Hash,
		PartSetHeader: types.CanonicalPartSetHeader{
			Hash:  blockID.PartSetHeader.Hash,
			Total: blockID.PartSetHeader.Total,
		},
	}
	gotCanonicalBlockIDEnc, err := gotCanonicalBlockID.Marshal()
	if err != nil {
		t.Fatalf("proposal BlockID cannot be packed: %v", err)
	}
	if !bytes.Equal(gotCanonicalBlockIDEnc, canonicalBlockIDEnc) {
		t.Errorf("proposal BlockID (%+v) does not match expected BlockID (%+v)", gotCanonicalBlockID, canonicalBlockID)
	}

	timeNode, err := proposalNode.LookupByString("Timestamp")
	if err != nil {
		t.Fatalf("proposal is missing Timestamp: %v", err)
	}
	timestamp, err := shared.PackTime(timeNode)
	if err != nil {
		t.Fatalf("proposal Timestamp cannot be packed: %v", err)
	}
	if !timestamp.Equal(testProposal.Timestamp) {
		t.Errorf("proposal Timestamp (%s) does not match expected Timestamp (%s)", timestamp.String(), testProposal.Timestamp.String())
	}

	chainIDNode, err := proposalNode.LookupByString("ChainID")
	if err != nil {
		t.Fatalf("proposal is missing ChainID: %v", err)
	}
	chainID, err := chainIDNode.AsString()
	if err != nil {
		t.Fatalf("proposal ChainID should be of type String: %v", err)
	}
	if chainID != testProposal.ChainID {
		t.Errorf("proposal ChainID (%s) does not match expected ChainID (%s)", chainID, testProposal.ChainID)
	}
}

func testProposalEncode(t *testing.T) {
	proposalWriter := new(bytes.Buffer)
	if err := proposal.Encode(proposalNode, proposalWriter); err != nil {
		t.Fatalf("unable to encode proposal into writer: %v", err)
	}
	encodedProposal := proposalWriter.Bytes()
	if !bytes.Equal(encodedProposal, testProposalEncoding) {
		t.Errorf("proposal encoding (%x) does not match the expected encoding (%x)", encodedProposal, testProposalEncoding)
	}
}
