//go:build ignore
// +build ignore

package main

// based on https://github.com/ipld/go-ipld-prime-proto/blob/master/gen/main.go

import (
	"fmt"
	"os"

	"github.com/ipld/go-ipld-prime/schema"
	gengo "github.com/ipld/go-ipld-prime/schema/gen/go"
)

const (
	pkgName = "dagcosmos"
)

func main() {
	// initialize a new type system
	ts := new(schema.TypeSystem)
	ts.Init()

	// accumulate the different types
	accumulateBasicTypes(ts)
	accumulateCryptoTypes(ts)
	accumulateChainTypes(ts)
	accumulateStateDataStructures(ts)

	// verify internal correctness of the types
	if errs := ts.ValidateGraph(); errs != nil {
		for _, err := range errs {
			fmt.Printf("- %s\n", err)
		}
		os.Exit(1)
	}

	// generate the code
	adjCfg := &gengo.AdjunctCfg{}
	gengo.Generate(".", pkgName, *ts, adjCfg)
}

func accumulateBasicTypes(ts *schema.TypeSystem) {
	/*
		# Uint is a non-negative integer
		type Uint bytes

		# The main purpose of HexBytes is to enable HEX-encoding for json/encoding.
		type HexBytes bytes

		# Address is a type alias of a slice of bytes
		# An address is calculated by hashing the public key using sha256
		# and truncating it to only use the first 20 bytes of the slice
		type Address HexBytes

		# Hash is a type alias of a slice of 32 bytes
		type Hash HexBytes

		# Time represents a unix timestamp with nanosecond granularity
		type Time struct {
			Seconds Uint
			Nanoseconds Uint
		}

		# Version captures the consensus rules for processing a block in the blockchain,
		# including all blockchain data structures and the rules of the application's
		# state transition machine.
		type Version struct {
			Block Uint
			App   Uint
		}
	*/
	// we could more explicitly type our links with SpawnLinkReference
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnInt("Int"))
	ts.Accumulate(schema.SpawnLink("Link"))
	ts.Accumulate(schema.SpawnBytes("Bytes"))
	ts.Accumulate(schema.SpawnBytes("Uint"))
	ts.Accumulate(schema.SpawnBytes("HexBytes"))
	ts.Accumulate(schema.SpawnBytes("Address"))
	ts.Accumulate(schema.SpawnBytes("Hash"))
	ts.Accumulate(schema.SpawnBytes("Duration"))
	ts.Accumulate(schema.SpawnStruct("Time",
		[]schema.StructField{
			schema.SpawnStructField("Seconds", "Uint", false, false),
			schema.SpawnStructField("Nanoseconds", "Uint", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("Version",
		[]schema.StructField{
			schema.SpawnStructField("Block", "Uint", false, false),
			schema.SpawnStructField("App", "Uint", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*	Params
		# ConsensusParams contains consensus critical parameters that determine the validity of blocks.
		type ConsensusParams struct {
			Block     BlockParams
			Evidence  EvidenceParams
			Validator ValidatorParams
			Version   VersionParams
		}

		# BlockParams contains limits on the block size and time between consecutive blocks
		type BlockParams struct {
			# Note: must be greater than 0
			MaxBytes Int
			# Note: must be greater or equal to -1
			MaxGas Int
		}

		# EvidenceParams determine how we handle evidence of malfeasance.
		type EvidenceParams struct {
			# Max age of evidence, in blocks.
			#
			# The basic formula for calculating this is: MaxAgeDuration / {average block
			# time}.
			MaxAgeNumBlocks Int
			# Max age of evidence, in time.
			#
			# It should correspond with an app's "unbonding period" or other similar
			# mechanism for handling [Nothing-At-Stake
			# attacks](https:#github.com/ethereum/wiki/wiki/Proof-of-Stake-FAQ#what-is-the-nothing-at-stake-problem-and-how-can-it-be-fixed).
			MaxAgeDuration time.Duration `protobuf:"bytes,2,opt,name=max_age_duration,json=maxAgeDuration,proto3,stdduration" json:"max_age_duration"`
			# This sets the maximum size of total evidence in bytes that can be committed in a single block.
			# and should fall comfortably under the max block bytes.
			# Default is 1048576 or 1MB
			MaxBytes Int
		}

		# ValidatorParams restrict the public key types validators can use.
		# NOTE: uses ABCI pubkey naming, not Amino names.
		type ValidatorParams struct {
			PubKeyTypes [String]
		}

		# VersionParams contains the ABCI application version.
		type VersionParams struct {
			AppVersion Uint
		}
	*/
	ts.Accumulate(schema.SpawnStruct("BlockParams",
		[]schema.StructField{
			schema.SpawnStructField("MaxBytes", "Int", false, false),
			schema.SpawnStructField("MaxGas", "Int", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("EvidenceParams",
		[]schema.StructField{
			schema.SpawnStructField("MaxAgeNumBlocks", "Int", false, false),
			schema.SpawnStructField("MaxAgeDuration", "Duration", false, false),
			schema.SpawnStructField("MaxBytes", "Int", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnList("PubKeyTypes", "String", false))
	ts.Accumulate(schema.SpawnStruct("ValidatorParams",
		[]schema.StructField{
			schema.SpawnStructField("PubKeyTypes", "PubKeyTypes", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("VersionParams",
		[]schema.StructField{
			schema.SpawnStructField("AppVersion", "Uint", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("ConsensusParams",
		[]schema.StructField{
			schema.SpawnStructField("Block", "BlockParams", false, false),
			schema.SpawnStructField("Evidence", "EvidenceParams", false, false),
			schema.SpawnStructField("Validator", "ValidatorParams", false, false),
			schema.SpawnStructField("Version", "VersionParams", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
}

func accumulateCryptoTypes(ts *schema.TypeSystem) {
	/*
		# Signatures in Tendermint are raw bytes representing the underlying signature
		type Signature bytes

		# Public key
		type PubKey bytes

		# Private key
		type PrivKey bytes

		# Proof represents a Merkle proof.
		# NOTE: The convention for proofs is to include leaf hashes but to
		# exclude the root hash.
		# This convention is implemented across IAVL range proofs as well.
		# Keep this consistent unless there's a very good reason to change
		# everything.  This also affects the generalized proof system as
		# well.
		type Proof struct {
			Total    Int
			Index    Int
			LeafHash Bytes
			Aunts    [Bytes]
		}
	*/
	ts.Accumulate(schema.SpawnBytes("Signature"))
	ts.Accumulate(schema.SpawnBytes("PubKey"))
	ts.Accumulate(schema.SpawnBytes("PrivKey"))
	ts.Accumulate(schema.SpawnList("Aunts", "Hash", false))
	ts.Accumulate(schema.SpawnStruct("Proof",
		[]schema.StructField{
			schema.SpawnStructField("Total", "Int", false, false),
			schema.SpawnStructField("Index", "Int", false, false),
			schema.SpawnStructField("LeafHash", "Hash", false, false),
			schema.SpawnStructField("Aunts", "Aunts", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
}

func accumulateChainTypes(ts *schema.TypeSystem) {
	/*
		# PartSetHeader is used for secure gossiping of the block during consensus
		# It contains the Merkle root of the complete serialized block cut into parts (ie. MerkleRoot(MakeParts(block))).
		type PartSetHeader struct {
			Total Uint
			Hash  PartTreeCID
		}

		# PartTreeCID is a CID link to the root node of a Part merkle tree
		# This CID is composed of the SHA_256 multihash of the root node in the Part merkle tree and the Part codec (tbd)
		# Part merkle tree is a Merkle tree built from the PartSet
		type PartTreeCID &MerkleTreeNode

		# PartSet is the complete set of parts for a header
		type PartSet [Part]

		# Part is a section of bytes of a complete serialized header
		type Part struct {
			Index Uint
			Bytes HexBytes
			Proof Proof
		}
	*/
	ts.Accumulate(schema.SpawnStruct("Part",
		[]schema.StructField{
			schema.SpawnStructField("Index", "Uint", false, false),
			schema.SpawnStructField("Bytes", "HexBytes", false, false),
			schema.SpawnStructField("Proof", "Proof", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnList("PartSet", "Part", false))
	ts.Accumulate(schema.SpawnStruct("PartSetHeader",
		[]schema.StructField{
			schema.SpawnStructField("Total", "Uint", false, false),
			schema.SpawnStructField("Hash", "Link", false, false), // link to the root node of a merkle tree created from part set
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
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
	*/
	ts.Accumulate(schema.SpawnStruct("BlockID",
		[]schema.StructField{
			schema.SpawnStructField("Hash", "Link", false, false), // HeaderCID, link to the root node of a merkle tree created from all the consensus fields in a header
			schema.SpawnStructField("PartSetHeader", "PartSetHeader", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("Header",
		[]schema.StructField{
			schema.SpawnStructField("Version", "Version", false, false),
			schema.SpawnStructField("ChainID", "String", false, false),
			schema.SpawnStructField("Height", "Int", false, false),
			schema.SpawnStructField("Time", "Time", false, false),
			schema.SpawnStructField("LastBlockID", "BlockID", false, false),
			schema.SpawnStructField("LastCommitHash", "Link", false, false),     // CommitTreeCID
			schema.SpawnStructField("DataHash", "Link", false, false),           // TxTreeCID
			schema.SpawnStructField("ValidatorsHash", "Link", false, false),     // ValidatorTreeCID
			schema.SpawnStructField("NextValidatorsHash", "Link", false, false), // ValidatorTreeCID
			schema.SpawnStructField("ConsensusHash", "Link", false, false),      // HashedParamsCID
			schema.SpawnStructField("AppHash", "Link", false, false),            // AppStateTreeCID
			schema.SpawnStructField("LastResultsHash", "Link", false, false),    // LastResultsHash
			schema.SpawnStructField("EvidenceHash", "Link", false, false),       // EvidenceTreeCID
			schema.SpawnStructField("ProposerAddress", "Address", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		type Tx [Bytes]

		type Data struct {
			Txs [Tx]
		}
	*/
	ts.Accumulate(schema.SpawnList("Tx", "Bytes", false))
	ts.Accumulate(schema.SpawnList("Txs", "Tx", false))
	ts.Accumulate(schema.SpawnStruct("Data",
		[]schema.StructField{
			schema.SpawnStructField("Txs", "Txs", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))

	/*
		# BlockIDFlag is a single byte flag
		type BlockIDFlag enum {
		  | BlockIDFlagUnknown ("0")
		  | BlockIDFlagAbsent ("1")
		  | BlockIDFlagCommit ("2")
		  | BlockIDFlagNil ("3")
		} representation int

		# CommitSig is a part of the Vote included in a Commit.
		type CommitSig struct {
			BlockIDFlag      BlockIDFlag
			ValidatorAddress Address
			Timestamp        Time
			Signature        Signature
		}

		type Signatures [CommitSig]

		# Commit contains
		type Commit struct {
			# NOTE: The signatures are in order of address to preserve the bonded
			# ValidatorSet order.
			# Any peer with a block can gossip signatures by index with a peer without
			# recalculating the active ValidatorSet.
			Height     Int
			Round      Int
			BlockID    BlockID
			Signatures []CommitSig
		}
	*/
	// make this an enum after schema gen support is added
	ts.Accumulate(schema.SpawnInt("BlockIDFlag"))
	ts.Accumulate(schema.SpawnStruct("CommitSig",
		[]schema.StructField{
			schema.SpawnStructField("BlockIDFlag", "BlockIDFlag", false, false),
			schema.SpawnStructField("ValidatorAddress", "Address", false, false),
			schema.SpawnStructField("Timestamp", "Time", false, false),
			schema.SpawnStructField("Signature", "Signature", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnList("Signatures", "CommitSig", false))
	ts.Accumulate(schema.SpawnStruct("Commit",
		[]schema.StructField{
			schema.SpawnStructField("Height", "Int", false, false),
			schema.SpawnStructField("Round", "Int", false, false),
			schema.SpawnStructField("BlockID", "BlockID", false, false),
			schema.SpawnStructField("Signatures", "Signatures", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		# Vote represents a prevote, precommit, or commit vote from validators for
		# consensus.
		type Vote struct {
			Type             SignedMsgType
			Height           Int
			Round            Int
			BlockID          BlockID
			Timestamp        Time
			ValidatorAddress Address
			ValidatorIndex   Int
			Signature        Signature
		}

		# SignedMsgType is the type of signed message in the consensus.
		type SignedMsgType enum {
			| UnknownType ("0")
			| PrevoteType ("1")
			| PrecommitType ("2")
			| ProposalType ("32")
		} representation int

		# Proposal defines a block proposal for the consensus.
		# It refers to the block by BlockID field.
		# It must be signed by the correct proposer for the given Height/Round
		# to be considered valid. It may depend on votes from a previous round,
		# a so-called Proof-of-Lock (POL) round, as noted in the POLRound.
		# If POLRound >= 0, then BlockID corresponds to the block that is locked in POLRound.
		type Proposal struct {
			Type      SignedMsgType
			Height    Int
			Round     Int # there can not be greater than 2_147_483_647 rounds
			POLRound  Int # -1 if null.
			BlockID   BlockID
			Timestamp Time
			Signature Signature
		}
	*/
	// make this an enum after schema gen support is added
	ts.Accumulate(schema.SpawnInt("SignedMsgType"))
	ts.Accumulate(schema.SpawnStruct("Vote",
		[]schema.StructField{
			schema.SpawnStructField("Type", "SignedMsgType", false, false),
			schema.SpawnStructField("Height", "Int", false, false),
			schema.SpawnStructField("Round", "Int", false, false),
			schema.SpawnStructField("BlockID", "BlockID", false, false),
			schema.SpawnStructField("Timestamp", "Time", false, false),
			schema.SpawnStructField("ValidatorAddress", "Address", false, false),
			schema.SpawnStructField("ValidatorIndex", "Int", false, false),
			schema.SpawnStructField("Signature", "Signature", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("Proposal",
		[]schema.StructField{
			schema.SpawnStructField("Type", "SignedMsgType", false, false),
			schema.SpawnStructField("Height", "Int", false, false),
			schema.SpawnStructField("Round", "Int", false, false),
			schema.SpawnStructField("POLRound", "Int", false, false),
			schema.SpawnStructField("BlockID", "BlockID", false, false),
			schema.SpawnStructField("Timestamp", "Time", false, false),
			schema.SpawnStructField("Signature", "Signature", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))

	/*
		# Volatile state for each Validator
		# NOTE: The Address and ProposerPriority is not included in Validator.Hash();
		# make sure to update that method if changes are made here
		type Validator struct {
			Address     Address # this should be remove since it isn't included in the content hahs?
			PubKey      PubKey
			VotingPower Int
			ProposerPriority Int # this should be removed since it isn't included in the content hash?
		}

		# This is what is actually hashed...
		type SimpleValidator struct {
			PubKey      PubKey
			VotingPower Int
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
	ts.Accumulate(schema.SpawnStruct("Validator",
		[]schema.StructField{
			schema.SpawnStructField("Address", "Address", false, false),
			schema.SpawnStructField("PubKey", "PubKey", false, false),
			schema.SpawnStructField("VotingPower", "Int", false, false),
			schema.SpawnStructField("ProsperPriority", "Int", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("SimpleValidator",
		[]schema.StructField{
			schema.SpawnStructField("PubKey", "PubKey", false, false),
			schema.SpawnStructField("VotingPower", "Int", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnList("Validators", "Validator", false))
	ts.Accumulate(schema.SpawnStruct("ValidatorSet",
		[]schema.StructField{
			schema.SpawnStructField("Validators", "Validators", false, false),
			schema.SpawnStructField("Proposer", "Validator", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))

	/*
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
	*/
	ts.Accumulate(schema.SpawnStruct("SignedHeader",
		[]schema.StructField{
			schema.SpawnStructField("Header", "Header", false, false),
			schema.SpawnStructField("Commit", "Commit", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("LightBlock",
		[]schema.StructField{
			schema.SpawnStructField("SignedHeader", "SignedHeader", false, false),
			schema.SpawnStructField("ValidatorSet", "ValidatorSet", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		# EvidenceData contains any evidence of malicious wrong-doing by validators
		type EvidenceData struct {
			Evidence EvidenceList
		}

		# EvidenceList is a list of Evidence
		type EvidenceList [Evidence]

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
	ts.Accumulate(schema.SpawnStruct("DuplicateVoteEvidence",
		[]schema.StructField{
			schema.SpawnStructField("VoteA", "Vote", false, false),
			schema.SpawnStructField("VoteB", "Vote", false, false),
			schema.SpawnStructField("TotalVotingPower", "Int", false, false),
			schema.SpawnStructField("ValidatorPower", "Int", false, false),
			schema.SpawnStructField("Timestamp", "Time", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))

	ts.Accumulate(schema.SpawnStruct("LightClientAttackEvidence",
		[]schema.StructField{
			schema.SpawnStructField("ConflictingBlock", "LightBlock", false, false),
			schema.SpawnStructField("CommonHeight", "Int", false, false),
			schema.SpawnStructField("ByzantineValidators", "Validators", false, false),
			schema.SpawnStructField("TotalVotingPower", "Int", false, false),
			schema.SpawnStructField("Timestamp", "Time", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnUnion("Evidence",
		[]schema.TypeName{
			"DuplicateVoteEvidence",
			"LightClientAttackEvidence",
		},
		schema.SpawnUnionRepresentationKeyed(map[string]schema.TypeName{
			"duplicate": "DuplicateVoteEvidence",
			"light":     "LightClientAttackEvidence",
		}),
	))
	ts.Accumulate(schema.SpawnList("EvidenceList", "Evidence", false))
	ts.Accumulate(schema.SpawnStruct("EvidenceData",
		[]schema.StructField{
			schema.SpawnStructField("Evidence", "EvidenceList", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		# HashedParams is a subset of ConsensusParams that is included in the consensus encoding
		# It is hashed into the Header.ConsensusHash.
		type HashedParams struct {
			BlockMaxBytes Int
			BlockMaxGas   Int
		}
	*/
	ts.Accumulate(schema.SpawnStruct("HashedParams",
		[]schema.StructField{
			schema.SpawnStructField("BlockMaxBytes", "Int", false, false),
			schema.SpawnStructField("BlockMaxGas", "Int", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		# Block isn't really an IPLD, its hash is a hash of just the header
		# Block defines the atomic unit of a Tendermint blockchain
		type Block struct {
			Header Header
			Data Data
			Evidence EvidenceData
			LastCommit Commit
		}
	*/
	ts.Accumulate(schema.SpawnStruct("Block",
		[]schema.StructField{
			schema.SpawnStructField("Header", "Header", false, false),
			schema.SpawnStructField("Data", "Data", false, false),
			schema.SpawnStructField("Evidence", "EvidenceData", false, false),
			schema.SpawnStructField("LastCommit", "Commit", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	/*
		type MerkleTreeNode union {
		    | MerkleTreeInnerNode "parent"
		    | MerkleTreeLeafNode "leaf"
		} representation keyed

		# MerkleTreeRootNode is the top-most parent node in a merkle tree; the root node of the tree.
		# It can be a leaf node if there is only one value in the tree
		type MerkleTreeRootNode MerkleTreeNode

		# MerkleTreeInnerNode nodes contain two byte arrays which contain the hashes which reference its two child nodes.
		type MerkleTreeInnerNode struct {
		    ChildA &MerkleTreeNode
		    ChildB &MerkleTreeNode
		}

		# MerkleTreeLeafNode is a single byte array containing the value stored at that leaf
		# Often times this "value" will be a hash of content rather than the content itself
		type MerkleTreeLeafNode struct {
		    Value Bytes
		}
	*/
	ts.Accumulate(schema.SpawnStruct("MerkleTreeLeafNode",
		[]schema.StructField{
			schema.SpawnStructField("Value", "Bytes", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnStruct("MerkleTreeInnerNode",
		[]schema.StructField{
			schema.SpawnStructField("ChildA", "Link", false, false),
			schema.SpawnStructField("ChildB", "Link", false, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
}

func accumulateStateDataStructures(ts *schema.TypeSystem) {
	// TODO: write schemas for state
}
