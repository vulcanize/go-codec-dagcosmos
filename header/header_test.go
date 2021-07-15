package header_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"

	dagcosmos "github.com/vulcanize/go-codec-dagcosmos"
	"github.com/vulcanize/go-codec-dagcosmos/header"
	"github.com/vulcanize/go-codec-dagcosmos/shared"
)

var (
	testHeader = types.Header{
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
	headerProto       = testHeader.ToProto()
	headerEncoding, _ = headerProto.Marshal()
	headerNode        ipld.Node
)

/* IPLD Schema
# Header defines the structure of a Tendermint block header
type Header struct {
	# basic block info
	Version Version
	ChainID String
	Height  Int
	Time    Time

	# prev block info
	LastBlockID BlockID

	# hashes of block data
	LastCommitHash CommitTreeCID
	DataHash       TxTreeCID

	# hashes from the app output from the prev block
	ValidatorsHash     ValidatorTreeCID # MerkleRoot of the current validator set
	NextValidatorsHash ValidatorTreeCID # MerkleRoot of the next validator set
	ConsensusHash      HashedParamsCID
	AppHash            AppStateTreeCID # State Root from the state machine

	# root hash of all results from the txs from the previous block
	LastResultsHash ResultTreeCID

	# consensus info
	EvidenceHash    EvidenceTreeCID
	ProposerAddress Address
}

# HashedParamsCID is a CID link to the HashedParams for this Header
# This CID is composed of the SHA_256 multihash of the linked protobuf encoded HashedParams struct and the HashedParmas codec (tbd)
type HashParamsCID &HashedParams

# EvidenceTreeCID is a CID link to the root node of a Evidence merkle tree
# This CID is composed of the SHA_256 multihash of the root node in the Evidence merkle tree and the Evidence codec (tbd)
# The Evidence merkle tree is Merkle tree build from the list of evidence of Byzantine behaviour included in this block.
type EvidenceTreeCID &MerkleTreeNode

# ResultTreeCID is a CID link to the root node of a Result merkle tree
# This CID is composed of the SHA_256 multihash of the root node in a Result merkle tree and the Result codec (tbd)
# Result merkle tree is a Merkle tree built from ResponseDeliverTx responses (Log, Info, Codespace and Events fields are ignored)
type ResultTreeCID &MerkleTreeNode

# AppStateTreeCID is a CID link to the state root returned by the state machine after executing and commiting the previous block
# It serves as the basis for validating any Merkle proofs that comes from the ABCI application and represents the state of the actual application rather than the state of the blockchain itself.
# This nature of the hash is determined by the application, Tendermint can not perform validation on it
type AppStateReference &MerkleTreeNode

# ValidatorTreeCID is a CID link to the root node of a Validator merkle tree
# This CID is composed of the SHA_256 multihash of the root node in the Validator merkle tree and the Validator codec (tbd)
# Validator merkle tree is a Merkle tree built from the set of validators for the given block
# The validators are first sorted by voting power (descending), then by address (ascending) prior to computing the MerkleRoot
type ValidatorTreeCID &MerkleTreeNode

# TxTreeCID is a CID link to the root node of a Tx merkle tree
# This CID is composed of the SHA_256 multihash of the root node in the Tx merkle tree and the Tx codec (tbd)
# Tx merkle tree is a Merkle tree built from the set of Txs at the given block
# Note: The transactions are hashed before being included in the Merkle tree, the leaves of the Merkle tree contain the hashes, not the transactions themselves.
type TxTreeCID &MerkleTreeNode

# CommitTreeCID is a CID link to the root node of a Commit merkle tree
# This CID is composed of the SHA_256 multihash of the root node in a Commit merkle tree and the Commit codec (tbd)
# Commit merkle tree is a Merkle tree built from a set of validator's commits
type CommitTreeCID &MerkleTreeNode

# BlockID contains two distinct Merkle roots of the block.
# The BlockID includes these two hashes, as well as the number of parts (ie. len(MakeParts(block)))
type BlockID struct {
	Hash          HeaderCID
	PartSetHeader PartSetHeader
}

# HeaderCID is a CID link to the root node of a Header merkle tree
# This CID is composed of the SHA_256 multihash of the root node in the Header merkle tree and the Header codec (tbd)
# Header merkle tree is a Merklization of all of the fields in the header
type HeaderCID &MerkleTreeNode

# Version captures the consensus rules for processing a block in the blockchain,
# including all blockchain data structures and the rules of the application's
# state transition machine.
type Version struct {
	Block Uint
	App   Uint
}
*/

func TestHeaderCodec(t *testing.T) {
	testHeaderDecode(t)
	testHeaderNodeContents(t)
	testHeaderEncode(t)
}

func testHeaderDecode(t *testing.T) {
	headerBuilder := dagcosmos.Type.Header.NewBuilder()
	headerReader := bytes.NewReader(headerEncoding)
	if err := header.Decode(headerBuilder, headerReader); err != nil {
		t.Fatalf("unable to decode header into an IPLD node: %v", err)
	}
	headerNode = headerBuilder.Build()
}

func testHeaderNodeContents(t *testing.T) {
	versionNode, err := headerNode.LookupByString("Version")
	if err != nil {
		t.Fatalf("header is missing Version: %v", err)
	}
	version, err := shared.PackVersion(versionNode)
	if err != nil {
		t.Fatalf("header Version cannot be packed: %v", err)
	}
	if version.App != testHeader.Version.App {
		t.Errorf("header App Version (%d) does not match expected App Version (%d)", version.App, testHeader.Version.App)
	}
	if version.Block != testHeader.Version.Block {
		t.Errorf("header Block Version (%d) does not match expected Block Version (%d)", version.Block, testHeader.Version.Block)
	}

	chainIDNode, err := headerNode.LookupByString("ChainID")
	if err != nil {
		t.Fatalf("header is missing ChainID: %v", err)
	}
	chainID, err := chainIDNode.AsString()
	if err != nil {
		t.Fatalf("header ChainID should be of type String: %v", err)
	}
	if chainID != testHeader.ChainID {
		t.Errorf("header ChainID (%s) does not match expected ChainID (%s)", chainID, testHeader.ChainID)
	}

	heightNode, err := headerNode.LookupByString("Height")
	if err != nil {
		t.Fatalf("header is missing Height: %v", err)
	}
	height, err := heightNode.AsInt()
	if err != nil {
		t.Fatalf("header Height should be of type Int: %v", err)
	}
	if height != testHeader.Height {
		t.Errorf("header Height (%d) does not match expected Height (%d)", height, testHeader.Height)
	}

	timeNode, err := headerNode.LookupByString("Time")
	if err != nil {
		t.Fatalf("header is missing Time: %v", err)
	}
	timestamp, err := shared.PackTime(timeNode)
	if err != nil {
		t.Fatalf("header Time cannot be packed: %v", err)
	}
	if !timestamp.Equal(testHeader.Time) {
		t.Errorf("header Time (%s) does not match expected Time (%s)", timestamp.String(), testHeader.Time.String())
	}

	lastBlockIDNode, err := headerNode.LookupByString("LastBlockID")
	if err != nil {
		t.Fatalf("header is missing LastBlockID: %v", err)
	}
	lastBlockID, err := shared.PackBlockID(lastBlockIDNode)
	if err != nil {
		t.Fatalf("header LastBlockID cannot be packed: %v", err)
	}
	if !lastBlockID.Equals(testHeader.LastBlockID) {
		t.Errorf("header LastBlockID (%s) does not match expected LastBlockID (%s)", lastBlockID.String(), testHeader.LastBlockID.String())
	}

	lchNode, err := headerNode.LookupByString("LastCommitHash")
	if err != nil {
		t.Fatalf("header is missing LastCommitHash: %v", err)
	}
	lch, err := shared.PackLink(lchNode)
	if err != nil {
		t.Fatalf("header LastCommitHash cannot be packed: %v", err)
	}
	if !bytes.Equal(lch, testHeader.LastCommitHash) {
		t.Errorf("header LastCommitHash (%x) does not match expected LastCommitHash (%x)", lch, testHeader.LastCommitHash)
	}

	dataNode, err := headerNode.LookupByString("DataHash")
	if err != nil {
		t.Fatalf("header is missing DataHash: %v", err)
	}
	data, err := shared.PackLink(dataNode)
	if err != nil {
		t.Fatalf("header DataHash cannot be packed: %v", err)
	}
	if !bytes.Equal(data, testHeader.DataHash) {
		t.Errorf("header DataHash (%x) does not match expected DataHash (%x)", data, testHeader.DataHash)
	}

	vhNode, err := headerNode.LookupByString("ValidatorsHash")
	if err != nil {
		t.Fatalf("header is missing ValidatorsHash: %v", err)
	}
	vh, err := shared.PackLink(vhNode)
	if err != nil {
		t.Fatalf("header ValidatorsHash cannot be packed: %v", err)
	}
	if !bytes.Equal(vh, testHeader.ValidatorsHash) {
		t.Errorf("header ValidatorsHash (%x) does not match expected ValidatorsHash (%x)", vh, testHeader.ValidatorsHash)
	}

	nvhNode, err := headerNode.LookupByString("NextValidatorsHash")
	if err != nil {
		t.Fatalf("header is missing NextValidatorsHash: %v", err)
	}
	nvh, err := shared.PackLink(nvhNode)
	if err != nil {
		t.Fatalf("header NextValidatorsHash cannot be packed: %v", err)
	}
	if !bytes.Equal(nvh, testHeader.NextValidatorsHash) {
		t.Errorf("header NextValidatorsHash (%x) does not match expected NextValidatorsHash (%x)", nvh, testHeader.NextValidatorsHash)
	}

	chNode, err := headerNode.LookupByString("ConsensusHash")
	if err != nil {
		t.Fatalf("header is missing ConsensusHash: %v", err)
	}
	ch, err := shared.PackLink(chNode)
	if err != nil {
		t.Fatalf("header ConsensusHash cannot be packed: %v", err)
	}
	if !bytes.Equal(ch, testHeader.ConsensusHash) {
		t.Errorf("header ConsensusHash (%x) does not match expected ConsensusHash (%x)", ch, testHeader.ConsensusHash)
	}

	appHashNode, err := headerNode.LookupByString("AppHash")
	if err != nil {
		t.Fatalf("header is missing AppHash: %v", err)
	}
	appHash, err := shared.PackLink(appHashNode)
	if err != nil {
		t.Fatalf("header AppHash cannot be packed: %v", err)
	}
	if !bytes.Equal(appHash, testHeader.AppHash) {
		t.Errorf("header AppHash (%x) does not match expected AppHash (%x)", appHash, testHeader.AppHash)
	}

	lrHashNode, err := headerNode.LookupByString("LastResultsHash")
	if err != nil {
		t.Fatalf("header is missing LastResultsHash: %v", err)
	}
	lrHash, err := shared.PackLink(lrHashNode)
	if err != nil {
		t.Fatalf("header LastResultsHash cannot be packed: %v", err)
	}
	if !bytes.Equal(lrHash, testHeader.LastResultsHash) {
		t.Errorf("header LastResultsHash (%x) does not match expected LastResultsHash (%x)", lrHash, testHeader.LastResultsHash)
	}

	evHashNode, err := headerNode.LookupByString("EvidenceHash")
	if err != nil {
		t.Fatalf("header is missing EvidenceHash: %v", err)
	}
	evHash, err := shared.PackLink(evHashNode)
	if err != nil {
		t.Fatalf("header EvidenceHash cannot be packed: %v", err)
	}
	if !bytes.Equal(evHash, testHeader.EvidenceHash) {
		t.Errorf("header EvidenceHash (%x) does not match expected EvidenceHash (%x)", evHash, testHeader.EvidenceHash)
	}

	pAddrNode, err := headerNode.LookupByString("ProposerAddress")
	if err != nil {
		t.Fatalf("header is missing ProposerAddress: %v", err)
	}
	pAddr, err := pAddrNode.AsBytes()
	if err != nil {
		t.Fatalf("header ProposerAddress should be of type Bytes: %v", err)
	}
	if !bytes.Equal(pAddr, testHeader.ProposerAddress) {
		t.Errorf("header ProposerAddress (%x) does not match expected ProposerAddress (%x)", pAddr, testHeader.ProposerAddress)
	}
}

func testHeaderEncode(t *testing.T) {
	headerWriter := new(bytes.Buffer)
	if err := header.Encode(headerNode, headerWriter); err != nil {
		t.Fatalf("unable to encode header into writer: %v", err)
	}
	encodedHeaderBytes := headerWriter.Bytes()
	if !bytes.Equal(encodedHeaderBytes, headerEncoding) {
		t.Errorf("header encoding (%x) does not match the expected encoding (%x)", encodedHeaderBytes, headerEncoding)
	}
}
