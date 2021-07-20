package evidence_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/evidence"
	"github.com/vulcanize/go-codec-dagcosmos/light_block"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	privKey1     = ed25519.GenPrivKey()
	pubKey1      = privKey1.PubKey()
	privKey2     = ed25519.GenPrivKey()
	pubKey2      = privKey2.PubKey()
	privKey3     = ed25519.GenPrivKey()
	pubKey3      = privKey3.PubKey()
	voteABlockID = types.BlockID{
		Hash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		PartSetHeader: types.PartSetHeader{
			Total: 1,
			Hash:  shared.RandomHash(),
		},
	}
	voteA = &types.Vote{
		Type:             1,
		Height:           137,
		Round:            4,
		BlockID:          voteABlockID,
		Timestamp:        time.Now(),
		ValidatorAddress: shared.RandomAddr(),
		ValidatorIndex:   2,
		Signature:        shared.RandomSig(),
	}
	voteAProto       = voteA.ToProto()
	voteAEncoding, _ = voteAProto.Marshal()
	voteBBlockID     = types.BlockID{
		Hash: []byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		PartSetHeader: types.PartSetHeader{
			Total: 2,
			Hash:  shared.RandomHash(),
		},
	}
	voteB = &types.Vote{
		Type:             2,
		Height:           137,
		Round:            4,
		BlockID:          voteBBlockID,
		Timestamp:        time.Now(),
		ValidatorAddress: shared.RandomAddr(),
		ValidatorIndex:   3,
		Signature:        shared.RandomSig(),
	}
	voteBProto         = voteB.ToProto()
	voteBEncoding, _   = voteBProto.Marshal()
	validatorsEncoding = [][]byte{
		validator1Encoding,
		validator2Encoding,
	}
	duplicateVoteEvidence = types.DuplicateVoteEvidence{
		VoteA:            voteA,
		VoteB:            voteB,
		TotalVotingPower: 1337,
		ValidatorPower:   1338,
		Timestamp:        time.Now(),
	}
	validator1 = &types.Validator{
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
	validatorHash = validatorSet.Hash()
	header        = &types.Header{
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
	blockHash = header.Hash()
	commit    = &types.Commit{
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
	signedHeader = &types.SignedHeader{
		Header: header,
		Commit: commit,
	}
	conflictingBlock = &types.LightBlock{
		SignedHeader: signedHeader,
		ValidatorSet: validatorSet,
	}
	conflictingBlockProto, _    = conflictingBlock.ToProto()
	conflictingBlockEncoding, _ = conflictingBlockProto.Marshal()
	lightClientAttackEvidence   = types.LightClientAttackEvidence{
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
	evidenceBuilder := dagcosmos.Type.Evidence.NewBuilder()
	evidenceReader := bytes.NewReader(duplicateVoteEvidenceEncoding)
	if err := evidence.Decode(evidenceBuilder, evidenceReader); err != nil {
		t.Fatalf("unable to decode duplicate vote evidence into an IPLD node: %v", err)
	}
	duplicateVoteEvidenceNode = evidenceBuilder.Build()
}

func testDuplicateVoteEvidenceNodeContents(t *testing.T) {
	dupEvidenceNode, err := duplicateVoteEvidenceNode.LookupByString(evidence.DUPLICATE_EVIDENCE.String())
	if err != nil {
		t.Fatalf("evidence union should be duplicate vote evidence: %v", err)
	}

	voteANode, err := dupEvidenceNode.LookupByString("VoteA")
	if err != nil {
		t.Fatalf("duplicate vote evidence is missing VoteA: %v", err)
	}
	gotVoteA, err := shared.PackVote(voteANode)
	if err != nil {
		t.Fatalf("duplicate vote evidence VoteA cannot be packed: %v", err)
	}
	gotVoteAProto := gotVoteA.ToProto()
	gotVoteAEncoding, err := gotVoteAProto.Marshal()
	if err != nil {
		t.Fatalf("duplicate vote evidence unable to encode VoteA: %v", err)
	}
	if !bytes.Equal(gotVoteAEncoding, voteAEncoding) {
		t.Errorf("duplicate vote evidence VoteA (%+v) does not match expected VoteA (%+v)", gotVoteA, voteA)
	}

	voteBNode, err := dupEvidenceNode.LookupByString("VoteB")
	if err != nil {
		t.Fatalf("duplicate vote evidence is missing VoteB: %v", err)
	}
	gotVoteB, err := shared.PackVote(voteBNode)
	if err != nil {
		t.Fatalf("duplicate vote evidence VoteB cannot be packed: %v", err)
	}
	gotVoteBProto := gotVoteB.ToProto()
	gotVoteBEncoding, err := gotVoteBProto.Marshal()
	if err != nil {
		t.Fatalf("duplicate vote evidence unable to encode VoteB: %v", err)
	}
	if !bytes.Equal(gotVoteBEncoding, voteBEncoding) {
		t.Errorf("duplicate vote evidence VoteB (%+v) does not match expected VoteB (%+v)", gotVoteB, voteB)
	}

	totalVotingPowerNode, err := dupEvidenceNode.LookupByString("TotalVotingPower")
	if err != nil {
		t.Fatalf("duplicate vote evidence is missing TotalVotingPower: %v", err)
	}
	tvp, err := totalVotingPowerNode.AsInt()
	if err != nil {
		t.Fatalf("duplicate vote evidence TotalVotingPower should be of type Int: %v", err)
	}
	if tvp != duplicateVoteEvidence.TotalVotingPower {
		t.Errorf("duplicate vote evidence TotalVotingPower (%d) does not match expected TotalVotingPower (%d)", tvp, duplicateVoteEvidence.TotalVotingPower)
	}

	validatorPowerNode, err := dupEvidenceNode.LookupByString("ValidatorPower")
	if err != nil {
		t.Fatalf("duplicate vote evidence is missing ValidatorPower: %v", err)
	}
	vp, err := validatorPowerNode.AsInt()
	if err != nil {
		t.Fatalf("duplicate vote evidence ValidatorPower should be of type Int: %v", err)
	}
	if vp != duplicateVoteEvidence.ValidatorPower {
		t.Errorf("duplicate vote evidence ValidatorPower (%d) does not match expected ValidatorPower (%d)", vp, duplicateVoteEvidence.ValidatorPower)
	}

	timestampNode, err := dupEvidenceNode.LookupByString("Timestamp")
	if err != nil {
		t.Fatalf("duplicate vote evidence is missing Timestamp: %v", err)
	}
	timestamp, err := shared.PackTime(timestampNode)
	if err != nil {
		t.Fatalf("duplicate vote evidence Timestamp cannot be packed: %v", err)
	}
	if !timestamp.Equal(duplicateVoteEvidence.Timestamp) {
		t.Errorf("duplicate vote evidence Timestamp (%s) does not match expected Timestamp (%s)", timestamp.String(), duplicateVoteEvidence.Timestamp.String())
	}
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
	evidenceBuilder := dagcosmos.Type.Evidence.NewBuilder()
	evidenceReader := bytes.NewReader(lightClientAttackEvidenceEncoding)
	if err := evidence.Decode(evidenceBuilder, evidenceReader); err != nil {
		t.Fatalf("unable to decode light client attack evidence into an IPLD node: %v", err)
	}
	lightClientAttackEvidenceNode = evidenceBuilder.Build()
}

func testLightClientAttackEvidenceNodeContents(t *testing.T) {
	lcaEvidenceNode, err := lightClientAttackEvidenceNode.LookupByString(evidence.LIGHT_EVIDENCE.String())
	if err != nil {
		t.Fatalf("evidence union should be light client attack evidence: %v", err)
	}

	conflictingBlockNode, err := lcaEvidenceNode.LookupByString("ConflictingBlock")
	if err != nil {
		t.Fatalf("light client attack evidence is missing ConflictingBlock: %v", err)
	}
	lightBlock := new(types.LightBlock)
	if err := light_block.EncodeLightBlock(lightBlock, conflictingBlockNode); err != nil {
		t.Fatalf("light client attack evidence ConflictingBlock cannot be packed: %v", err)
	}
	lightBlockProto, err := lightBlock.ToProto()
	if err != nil {
		t.Fatalf("light client attack evidence ConflictingBlock cannot be converted to proto type: %v", err)
	}
	lightBlockEncoding, err := lightBlockProto.Marshal()
	if err != nil {
		t.Fatalf("light client attack evidence ConflictingBlock proto type cannot be marshalled: %v", err)
	}
	if !bytes.Equal(lightBlockEncoding, conflictingBlockEncoding) {
		t.Errorf("light client attack evidence ConflictingBlock (%+v) does not match expected (%+v)", *lightBlock, *conflictingBlock)
	}

	commonHeightNode, err := lcaEvidenceNode.LookupByString("CommonHeight")
	if err != nil {
		t.Fatalf("light client attack evidence is missing CommonHeight: %v", err)
	}
	commonHeight, err := commonHeightNode.AsInt()
	if err != nil {
		t.Fatalf("light client attack evidence CommonHeight should be of type Int: %v", err)
	}
	if commonHeight != lightClientAttackEvidence.CommonHeight {
		t.Errorf("light client attack evidence CommonHeight (%d) does not match expected CommonHeight (%d)", commonHeight, lightClientAttackEvidence.CommonHeight)
	}

	bvsNode, err := lcaEvidenceNode.LookupByString("ByzantineValidators")
	if err != nil {
		t.Fatalf("light client attack evidence is missing ByzantineValidators: %v", err)
	}
	if bvsNode.Length() != int64(len(validators)) {
		t.Fatalf("light client attack evidence number of ByzantineValidators (%d) does not match expected number of ByzantineValidators (%d)", bvsNode.Length(), len(validators))
	}
	bvsIT := bvsNode.ListIterator()
	for !bvsIT.Done() {
		i, validatorNode, err := bvsIT.Next()
		if err != nil {
			t.Fatalf("light client attack evidence ByzantineValidators iterator error: %v", err)
		}
		validator, err := shared.PackValidator(validatorNode)
		if err != nil {
			t.Fatalf("light client attack evidence ByzantineValidator cannot be packed: %v", err)

		}
		validatorProto, err := validator.ToProto()
		if err != nil {
			t.Fatalf("light client attack evidence ByzantineValidator cannot be converted to proto type: %v", err)
		}
		validatorEncoding, err := validatorProto.Marshal()
		if err != nil {
			t.Fatalf("light client attack evidence ByzantineValidator proto type cannot be marshalled: %v", err)
		}
		if !bytes.Equal(validatorEncoding, validatorsEncoding[i]) {
			t.Errorf("light client attack evidence ByzantineValidator (%+v) does not match expected ByzantineValidator (%+v)", *validator, *validators[i])
		}
	}

	tvpNode, err := lcaEvidenceNode.LookupByString("TotalVotingPower")
	if err != nil {
		t.Fatalf("light client attack evidence is missing TotalVotingPower: %v", err)
	}
	tvp, err := tvpNode.AsInt()
	if err != nil {
		t.Fatalf("light client attack evidence TotalVotingPower should be of type Int: %v", err)
	}
	if tvp != lightClientAttackEvidence.TotalVotingPower {
		t.Errorf("light client attack evidence TotalVotingPower (%d) does not match expected TotalVotingPower (%d)", tvp, lightClientAttackEvidence.TotalVotingPower)
	}

	timestampNode, err := lcaEvidenceNode.LookupByString("Timestamp")
	if err != nil {
		t.Fatalf("light client attack evidence is missing Timestamp: %v", err)
	}
	timestamp, err := shared.PackTime(timestampNode)
	if err != nil {
		t.Fatalf("light client attack evidence Timestamp cannot be packed: %v", err)
	}
	if !timestamp.Equal(lightClientAttackEvidence.Timestamp) {
		t.Errorf("light client attack evidence Timestamp (%s) does not match expected Timestamp (%s)", timestamp.String(), lightClientAttackEvidence.Timestamp.String())
	}
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
