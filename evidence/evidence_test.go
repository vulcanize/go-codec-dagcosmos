package evidence_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"

	"github.com/vulcanize/go-codec-dagcosmos/evidence"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	privKey1 = ed25519.GenPrivKey()
	pubKey1  = privKey1.PubKey()
	privKey2 = ed25519.GenPrivKey()
	pubKey2  = privKey2.PubKey()
	privKey3 = ed25519.GenPrivKey()
	pubKey3  = privKey3.PubKey()
	voteA    = &types.Vote{
		Type:             1,
		Height:           137,
		Round:            4,
		BlockID:          types.BlockID{},
		Timestamp:        time.Time{},
		ValidatorAddress: shared.RandomAddr(),
		ValidatorIndex:   2,
		Signature:        shared.RandomSig(),
	}
	voteB = &types.Vote{
		Type:             32,
		Height:           137,
		Round:            4,
		BlockID:          types.BlockID{},
		Timestamp:        time.Time{},
		ValidatorAddress: shared.RandomAddr(),
		ValidatorIndex:   3,
		Signature:        shared.RandomSig(),
	}
	duplicateVoteEvidence = types.DuplicateVoteEvidence{
		VoteA:            voteA,
		VoteB:            voteB,
		TotalVotingPower: 1337,
		ValidatorPower:   1338,
		Timestamp:        time.Now(),
	}
	validators = []*types.Validator{
		&types.Validator{
			Address:          shared.RandomAddr(),
			PubKey:           pubKey1,
			VotingPower:      13337,
			ProposerPriority: 13338,
		},
		&types.Validator{
			Address:          shared.RandomAddr(),
			PubKey:           pubKey2,
			VotingPower:      133337,
			ProposerPriority: 133338,
		},
	}
	proposer = &types.Validator{
		Address:          shared.RandomAddr(),
		PubKey:           pubKey3,
		VotingPower:      1333337,
		ProposerPriority: 1333338,
	}
	validatorSet = &types.ValidatorSet{
		Validators: validators,
		Proposer:   proposer,
	}
	header = &types.Header{
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
		ValidatorsHash:     shared.RandomHash(),
		NextValidatorsHash: shared.RandomHash(),
		ConsensusHash:      shared.RandomHash(),
		AppHash:            shared.RandomHash(),
		LastResultsHash:    shared.RandomHash(),
		EvidenceHash:       shared.RandomHash(),
		ProposerAddress:    shared.RandomAddr(),
	}
	signatures = []types.CommitSig{
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
	commit = &types.Commit{
		Height: 137,
		Round:  4,
		BlockID: types.BlockID{
			Hash: shared.RandomHash(),
			PartSetHeader: types.PartSetHeader{
				Hash:  shared.RandomHash(),
				Total: 13333337,
			},
		},
		Signatures: signatures,
	}
	signedHeader = &types.SignedHeader{
		Header: header,
		Commit: commit,
	}
	conflictingBlock = &types.LightBlock{
		SignedHeader: signedHeader,
		ValidatorSet: validatorSet,
	}
	lightClientAttackEvidence = types.LightClientAttackEvidence{
		ConflictingBlock:    conflictingBlock,
		CommonHeight:        1,
		ByzantineValidators: validators,
		TotalVotingPower:    1337,
		Timestamp:           time.Now(),
	}
	duplicateVoteEvidenceProto        = duplicateVoteEvidence.ToProto()
	lightClientAttackEvidenceProto, _ = lightClientAttackEvidence.ToProto()

	duplicateVoteEvidenceEncoding, _                         = duplicateVoteEvidenceProto.Marshal()
	lightClientAttackEvidenceEncoding, _                     = lightClientAttackEvidenceProto.Marshal()
	duplicateVoteEvidenceNode, lightClientAttackEvidenceNode ipld.Node
)

/* IPLD Schema
# DuplicateVoteEvidence contains evidence of a single validator signing two conflicting votes.
type DuplicateVoteEvidence struct {
	VoteA Vote
	VoteB Vote

	# abci specific information
	TotalVotingPower Int
	ValidatorPower   Int
	Timestamp        Time
}

# LightClientAttackEvidence is a generalized evidence that captures all forms of known attacks on
# a light client such that a full node can verify, propose and commit the evidence on-chain for
# punishment of the malicious validators. There are three forms of attacks: Lunatic, Equivocation and Amnesia.
type LightClientAttackEvidence struct {
	ConflictingBlock LightBlock
	CommonHeight     Int

	# abci specific information
	ByzantineValidators [Validator] # validators in the validator set that misbehaved in creating the conflicting block
	TotalVotingPower    Int        # total voting power of the validator set at the common height
	Timestamp           Time         # timestamp of the block at the common height
}

# Evidence in Tendermint is used to indicate breaches in the consensus by a validator
type Evidence union {
  | DuplicateVoteEvidence "duplicate"
  | LightClientAttackEvidence "light"
} representation keyed
*/

func TestEvidenceCodec(t *testing.T) {
	testDuplicateVoteEvidenceDecode(t)
	testDuplicateVoteEvidenceNodeContents(t)
	testDuplicateVoteEvidenceEncode(t)
	testLightClientAttackEvidenceDecode(t)
	testLightClientAttackEvidenceNodeContents(t)
	testLightClientAttackEvidenceEncode(t)
}

func testDuplicateVoteEvidenceDecode(t *testing.T) {

}

func testDuplicateVoteEvidenceNodeContents(t *testing.T) {

}

func testDuplicateVoteEvidenceEncode(t *testing.T) {
	evidenceWriter := new(bytes.Buffer)
	if err := evidence.Encode(duplicateVoteEvidenceNode, evidenceWriter); err != nil {
		t.Fatalf("unable to encode duplicate vote evidence into writer: %v", err)
	}
	encodedDuplicateVoteEvidence := evidenceWriter.Bytes()
	if !bytes.Equal(encodedDuplicateVoteEvidence, duplicateVoteEvidenceEncoding) {
		t.Errorf("duplicate vote evidence encoding (%x) does not match the expected encoding (%x)", encodedDuplicateVoteEvidence, duplicateVoteEvidenceEncoding)
	}
}

func testLightClientAttackEvidenceDecode(t *testing.T) {

}

func testLightClientAttackEvidenceNodeContents(t *testing.T) {

}

func testLightClientAttackEvidenceEncode(t *testing.T) {
	evidenceWriter := new(bytes.Buffer)
	if err := evidence.Encode(lightClientAttackEvidenceNode, evidenceWriter); err != nil {
		t.Fatalf("unable to encode light client attack evidence into writer: %v", err)
	}
	encodedLightClientAttackEvidence := evidenceWriter.Bytes()
	if !bytes.Equal(encodedLightClientAttackEvidence, lightClientAttackEvidenceEncoding) {
		t.Errorf("light client attack evidence encoding (%x) does not match the expected encoding (%x)", encodedLightClientAttackEvidence, lightClientAttackEvidenceEncoding)
	}
}
