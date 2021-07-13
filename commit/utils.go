package commit

import (
	"github.com/ipld/go-ipld-prime"
	"github.com/tendermint/tendermint/types"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

// PackCommit packs a Commit from the provided ipld.Node
func PackCommit(commitNode ipld.Node) (*types.Commit, error) {
	heightNode, err := commitNode.LookupByString("Height")
	if err != nil {
		return nil, err
	}
	height, err := heightNode.AsInt()
	if err != nil {
		return nil, err
	}
	roundNode, err := commitNode.LookupByString("Round")
	if err != nil {
		return nil, err
	}
	round, err := roundNode.AsInt()
	if err != nil {
		return nil, err
	}
	blockIDNode, err := commitNode.LookupByString("BlockID")
	blockID, err := shared.PackBlockID(blockIDNode)
	if err != nil {
		return nil, err
	}
	signaturesNode, err := commitNode.LookupByString("Signatures")
	if err != nil {
		return nil, err
	}
	signatures := make([]types.CommitSig, signaturesNode.Length())
	signatureLI := signaturesNode.ListIterator()
	for !signatureLI.Done() {
		i, commitSigNode, err := signatureLI.Next()
		if err != nil {
			return nil, err
		}
		commitSig := new(types.CommitSig)
		if err := EncodeCommitSig(commitSig, commitSigNode); err != nil {
			return nil, err
		}
		signatures[i] = *commitSig
	}
	return &types.Commit{
		Height:     height,
		Round:      int32(round),
		BlockID:    blockID,
		Signatures: signatures,
	}, nil
}

// UnpackCommit unpacks Commit into NodeAssembler
func UnpackCommit(cma ipld.MapAssembler, c types.Commit) error {
	if err := cma.AssembleKey().AssignString("Height"); err != nil {
		return err
	}
	if err := cma.AssembleValue().AssignInt(c.Height); err != nil {
		return err
	}
	if err := cma.AssembleKey().AssignString("Round"); err != nil {
		return err
	}
	if err := cma.AssembleValue().AssignInt(int64(c.Round)); err != nil {
		return err
	}
	if err := cma.AssembleKey().AssignString("BlockID"); err != nil {
		return err
	}
	bidMA, err := cma.AssembleValue().BeginMap(2)
	if err != nil {
		return err
	}
	if err := shared.UnpackBlockID(bidMA, c.BlockID); err != nil {
		return err
	}
	if err := cma.AssembleKey().AssignString("Signatures"); err != nil {
		return err
	}
	sigsLA, err := cma.AssembleValue().BeginList(int64(len(c.Signatures)))
	if err != nil {
		return err
	}
	for _, commitSig := range c.Signatures {
		if err := DecodeCommitSig(sigsLA.AssembleValue(), commitSig); err != nil {
			return err
		}
	}
	if err := sigsLA.Finish(); err != nil {
		return err
	}
	return cma.Finish()
}
