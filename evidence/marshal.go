package evidence

import (
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/light_block"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

type EvidenceKind string

const (
	LIGHT_EVIDENCE     EvidenceKind = "light"
	DUPLICATE_EVIDENCE EvidenceKind = "duplicate"
)

func (n EvidenceKind) String() string {
	return string(n)
}

// Encode provides an IPLD codec encode interface for Tendermint Evidence IPLDs.
// This function is registered via the go-ipld-prime link loader for multicodec
// code XXX when this package is invoked via init.
func Encode(node ipld.Node, w io.Writer) error {
	// 1KiB can be allocated on the stack, and covers most small nodes
	// without having to grow the buffer and cause allocations.
	enc := make([]byte, 0, 1024)

	enc, err := AppendEncode(enc, node)
	if err != nil {
		return err
	}
	_, err = w.Write(enc)
	return err
}

// AppendEncode is like Encode, but it uses a destination buffer directly.
// This means less copying of bytes, and if the destination has enough capacity,
// fewer allocations.
func AppendEncode(enc []byte, inNode ipld.Node) ([]byte, error) {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.Evidence.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return nil, err
	}
	node := builder.Build()
	lightEvidenceNode, err := node.LookupByString(LIGHT_EVIDENCE.String())
	if err == nil {
		le := new(types.LightClientAttackEvidence)
		if err := EncodeLightEvidence(le, lightEvidenceNode); err != nil {
			return nil, err
		}
		tmle, err := le.ToProto()
		if err != nil {
			return nil, err
		}
		enc, err = tmle.Marshal()
		return enc, err
	}
	dupEvidenceNode, err := node.LookupByString(DUPLICATE_EVIDENCE.String())
	if err == nil {
		de := new(types.DuplicateVoteEvidence)
		if err := EncodeDuplicateEvidence(de, dupEvidenceNode); err != nil {
			return nil, err
		}
		tmde := de.ToProto()
		enc, err = tmde.Marshal()
		return enc, err
	}
	return nil, fmt.Errorf("unrecognized Evidence type")
}

// EncodeLightEvidence is like Encode, but it uses an input tendermint LightClientAttackEvidence type
func EncodeLightEvidence(le *types.LightClientAttackEvidence, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.LightClientAttackEvidence.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range leRequiredPackFuncs {
		if err := pFunc(le, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos CommitSig form (%v)", err)
		}
	}
	return nil
}

// EncodeDuplicateEvidence is like Encode, but it uses an input tendermint DuplicateVoteEvidence type
func EncodeDuplicateEvidence(de *types.DuplicateVoteEvidence, inNode ipld.Node) error {
	// Wrap in a typed node for some basic schema form checking
	builder := dagcosmos.Type.DuplicateVoteEvidence.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()
	for _, pFunc := range deRequiredPackFuncs {
		if err := pFunc(de, node); err != nil {
			return fmt.Errorf("invalid DAG-Cosmos CommitSig form (%v)", err)
		}
	}
	return nil
}

var deRequiredPackFuncs = []func(*types.DuplicateVoteEvidence, ipld.Node) error{
	packVoteA,
	packVoteB,
	packDETotalVotingPower,
	packValidatorPower,
	packDETimestamp,
}

var leRequiredPackFuncs = []func(*types.LightClientAttackEvidence, ipld.Node) error{
	packConflictingBlock,
	packCommonHeight,
	packByzantineValidators,
	packLETotalVotingPower,
	packLETimestamp,
}

func packVoteA(de *types.DuplicateVoteEvidence, node ipld.Node) error {
	voteNode, err := node.LookupByString("VoteA")
	if err != nil {
		return err
	}
	vote, err := shared.PackVote(voteNode)
	if err != nil {
		return err
	}
	de.VoteA = vote
	return nil
}

func packVoteB(de *types.DuplicateVoteEvidence, node ipld.Node) error {
	voteNode, err := node.LookupByString("VoteB")
	if err != nil {
		return err
	}
	vote, err := shared.PackVote(voteNode)
	if err != nil {
		return err
	}
	de.VoteB = vote
	return nil
}

func packDETotalVotingPower(de *types.DuplicateVoteEvidence, node ipld.Node) error {
	tvp, err := packTotalVotingPower(node)
	if err != nil {
		return err
	}
	de.TotalVotingPower = tvp
	return nil
}

func packLETotalVotingPower(le *types.LightClientAttackEvidence, node ipld.Node) error {
	tvp, err := packTotalVotingPower(node)
	if err != nil {
		return err
	}
	le.TotalVotingPower = tvp
	return nil
}

func packTotalVotingPower(node ipld.Node) (int64, error) {
	tvpNode, err := node.LookupByString("TotalVotingPower")
	if err != nil {
		return 0, err
	}
	return tvpNode.AsInt()
}

func packValidatorPower(de *types.DuplicateVoteEvidence, node ipld.Node) error {
	vpNode, err := node.LookupByString("ValidatorPower")
	if err != nil {
		return err
	}
	vp, err := vpNode.AsInt()
	if err != nil {
		return err
	}
	de.ValidatorPower = vp
	return nil
}

func packDETimestamp(de *types.DuplicateVoteEvidence, node ipld.Node) error {
	timeNode, err := node.LookupByString("Timestamp")
	if err != nil {
		return err
	}
	time, err := shared.PackTime(timeNode)
	if err != nil {
		return err
	}
	de.Timestamp = time
	return nil
}

func packLETimestamp(le *types.LightClientAttackEvidence, node ipld.Node) error {
	timeNode, err := node.LookupByString("Timestamp")
	if err != nil {
		return err
	}
	time, err := shared.PackTime(timeNode)
	if err != nil {
		return err
	}
	le.Timestamp = time
	return nil
}

func packConflictingBlock(le *types.LightClientAttackEvidence, node ipld.Node) error {
	lbNode, err := node.LookupByString("ConflictingBlock")
	if err != nil {
		return err
	}
	lb := new(types.LightBlock)
	if err := light_block.EncodeLightBlock(lb, lbNode); err != nil {
		return err
	}
	le.ConflictingBlock = lb
	return nil
}

func packCommonHeight(le *types.LightClientAttackEvidence, node ipld.Node) error {
	chNode, err := node.LookupByString("CommonHeight")
	if err != nil {
		return err
	}
	ch, err := chNode.AsInt()
	if err != nil {
		return err
	}
	le.CommonHeight = ch
	return nil
}

func packByzantineValidators(le *types.LightClientAttackEvidence, node ipld.Node) error {
	byzValsNode, err := node.LookupByString("ByzantineValidators")
	if err != nil {
		return err
	}
	byzVals := make([]*types.Validator, byzValsNode.Length())
	byzValsIT := byzValsNode.ListIterator()
	for !byzValsIT.Done() {
		i, validatorNode, err := byzValsIT.Next()
		if err != nil {
			return err
		}
		validator, err := shared.PackValidator(validatorNode)
		if err != nil {
			return err
		}
		byzVals[i] = validator
	}
	le.ByzantineValidators = byzVals
	return nil
}
