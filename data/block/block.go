package block

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data"
)

// This file holds the data structures related with the functionality of a shard block
//
// MiniBlock structure represents the body of a transaction block, holding an array of miniblocks
// each of the miniblocks has a different destination shard
// The body can be transmitted even before having built the heder and go through a prevalidation of each transaction

// Type identifies the type of the block
type Type uint8

// Body should be used when referring to the full list of mini blocks that forms a block body
type Body []*MiniBlock

// MiniBlockSlice should be used when referring to subset of mini blocks that is not
//  necessarily representing a full block body
type MiniBlockSlice []*MiniBlock

const (
	// TxBlock identifies a miniblock holding transactions
	TxBlock Type = 0
	// StateBlock identifies a miniblock holding account state
	StateBlock Type = 30
	// PeerBlock identifies a miniblock holding peer assignation
	PeerBlock Type = 60
	// SmartContractResultBlock identifies a miniblock holding smartcontractresults
	SmartContractResultBlock Type = 90
	// InvalidBlock identifies a miniblock holding invalid transactions
	InvalidBlock Type = 120
	// ReceiptBlock identifies a miniblock holding receipts
	ReceiptBlock Type = 150
	// TODO: leave rewards with highest value

	// RewardsBlock identifies a miniblock holding accumulated rewards, both system generated and from tx fees
	RewardsBlock Type = 255
)

// String returns the string representation of the Type
func (bType Type) String() string {
	switch bType {
	case TxBlock:
		return "TxBody"
	case StateBlock:
		return "StateBody"
	case PeerBlock:
		return "PeerBody"
	case SmartContractResultBlock:
		return "SmartContractResultBody"
	case RewardsBlock:
		return "RewardsBody"
	case InvalidBlock:
		return "InvalidBlock"
	case ReceiptBlock:
		return "ReceiptBlock"
	default:
		return fmt.Sprintf("Unknown(%d)", bType)
	}
}

// MiniBlock holds the transactions and the sender/destination shard ids
type MiniBlock struct {
	TxHashes        [][]byte
	ReceiverShardID uint32
	SenderShardID   uint32
	Type            Type
}

// MiniBlockHeader holds the hash of a miniblock together with sender/deastination shard id pair.
// The shard ids are both kept in order to differentiate between cross and single shard transactions
type MiniBlockHeader struct {
	Hash            []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	TxCount         uint32
	Type            Type
}

// PeerChange holds a change in one peer to shard assignation
type PeerChange struct {
	PubKey      []byte
	ShardIdDest uint32
}

// Header holds the metadata of a block. This is the part that is being hashed and run through consensus.
// The header holds the hash of the body and also the link to the previous block header hash
type Header struct {
	Nonce                  uint64
	PrevHash               []byte
	PrevRandSeed           []byte
	RandSeed               []byte
	PubKeysBitmap          []byte
	TimeStamp              uint64
	Round                  uint64
	Signature              []byte
	LeaderSignature        []byte
	RootHash               []byte
	ValidatorStatsRootHash []byte
	MetaBlockHashes        [][]byte
	EpochStartMetaHash     []byte
	ReceiptsHash           []byte
	ChainID                []byte
	MiniBlockHeaders       []MiniBlockHeader
	PeerChanges            []PeerChange
	Epoch                  uint32
	TxCount                uint32
	ShardId                uint32
	BlockBodyType          Type
}

// GetShardID returns header shard id
func (h *Header) GetShardID() uint32 {
	return h.ShardId
}

// GetNonce returns header nonce
func (h *Header) GetNonce() uint64 {
	return h.Nonce
}

// GetEpoch returns header epoch
func (h *Header) GetEpoch() uint32 {
	return h.Epoch
}

// GetRound returns round from header
func (h *Header) GetRound() uint64 {
	return h.Round
}

// GetRootHash returns the roothash from header
func (h *Header) GetRootHash() []byte {
	return h.RootHash
}

// GetValidatorStatsRootHash returns the root hash for the validator statistics trie at this current block
func (h *Header) GetValidatorStatsRootHash() []byte {
	return h.ValidatorStatsRootHash
}

// GetPrevHash returns previous block header hash
func (h *Header) GetPrevHash() []byte {
	return h.PrevHash
}

// GetPrevRandSeed returns previous random seed
func (h *Header) GetPrevRandSeed() []byte {
	return h.PrevRandSeed
}

// GetRandSeed returns the random seed
func (h *Header) GetRandSeed() []byte {
	return h.RandSeed
}

// GetPubKeysBitmap return signers bitmap
func (h *Header) GetPubKeysBitmap() []byte {
	return h.PubKeysBitmap
}

// GetSignature returns signed data
func (h *Header) GetSignature() []byte {
	return h.Signature
}

// GetLeaderSignature returns the leader's signature
func (h *Header) GetLeaderSignature() []byte {
	return h.LeaderSignature
}

// GetChainID gets the chain ID on which this block is valid on
func (h *Header) GetChainID() []byte {
	return h.ChainID
}

// GetTimeStamp returns the time stamp
func (h *Header) GetTimeStamp() uint64 {
	return h.TimeStamp
}

// GetTxCount returns transaction count in the block associated with this header
func (h *Header) GetTxCount() uint32 {
	return h.TxCount
}

// SetShardID sets header shard ID
func (h *Header) SetShardID(shId uint32) {
	h.ShardId = shId
}

// GetReceiptsHash returns the hash of the receipts and intra-shard smart contract results
func (h *Header) GetReceiptsHash() []byte {
	return h.ReceiptsHash
}

// SetNonce sets header nonce
func (h *Header) SetNonce(n uint64) {
	h.Nonce = n
}

// SetEpoch sets header epoch
func (h *Header) SetEpoch(e uint32) {
	h.Epoch = e
}

// SetRound sets header round
func (h *Header) SetRound(r uint64) {
	h.Round = r
}

// SetRootHash sets root hash
func (h *Header) SetRootHash(rHash []byte) {
	h.RootHash = rHash
}

// SetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (h *Header) SetValidatorStatsRootHash(rHash []byte) {
	h.ValidatorStatsRootHash = rHash
}

// SetPrevHash sets prev hash
func (h *Header) SetPrevHash(pvHash []byte) {
	h.PrevHash = pvHash
}

// SetPrevRandSeed sets previous random seed
func (h *Header) SetPrevRandSeed(pvRandSeed []byte) {
	h.PrevRandSeed = pvRandSeed
}

// SetRandSeed sets previous random seed
func (h *Header) SetRandSeed(randSeed []byte) {
	h.RandSeed = randSeed
}

// SetPubKeysBitmap sets publick key bitmap
func (h *Header) SetPubKeysBitmap(pkbm []byte) {
	h.PubKeysBitmap = pkbm
}

// SetSignature sets header signature
func (h *Header) SetSignature(sg []byte) {
	h.Signature = sg
}

// SetLeaderSignature will set the leader's signature
func (h *Header) SetLeaderSignature(sg []byte) {
	h.LeaderSignature = sg
}

// SetChainID sets the chain ID on which this block is valid on
func (h *Header) SetChainID(chainID []byte) {
	h.ChainID = chainID
}

// SetTimeStamp sets header timestamp
func (h *Header) SetTimeStamp(ts uint64) {
	h.TimeStamp = ts
}

// SetTxCount sets the transaction count of the block associated with this header
func (h *Header) SetTxCount(txCount uint32) {
	h.TxCount = txCount
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (h *Header) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	hashDst := make(map[string]uint32)
	for _, val := range h.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}
	return hashDst
}

// MapMiniBlockHashesToShards is a map of mini block hashes and sender IDs
func (h *Header) MapMiniBlockHashesToShards() map[string]uint32 {
	hashDst := make(map[string]uint32)
	for _, val := range h.MiniBlockHeaders {
		hashDst[string(val.Hash)] = val.SenderShardID
	}
	return hashDst
}

// Clone returns a clone of the object
func (h *Header) Clone() data.HeaderHandler {
	headerCopy := *h
	return &headerCopy
}

// IntegrityAndValidity checks if data is valid
func (b Body) IntegrityAndValidity() error {
	if b == nil || b.IsInterfaceNil() {
		return data.ErrNilBlockBody
	}

	for i := 0; i < len(b); i++ {
		if len(b[i].TxHashes) == 0 {
			return data.ErrMiniBlockEmpty
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (b Body) IsInterfaceNil() bool {
	return b == nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *Header) IsInterfaceNil() bool {
	if h == nil {
		return true
	}
	return false
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (h *Header) IsStartOfEpochBlock() bool {
	return len(h.EpochStartMetaHash) > 0
}

// ItemsInHeader gets the number of items(hashes) added in block header
func (h *Header) ItemsInHeader() uint32 {
	itemsInHeader := len(h.MiniBlockHeaders) + len(h.PeerChanges) + len(h.MetaBlockHashes)
	return uint32(itemsInHeader)
}

// ItemsInBody gets the number of items(hashes) added in block body
func (h *Header) ItemsInBody() uint32 {
	return h.TxCount
}

// CheckChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (h *Header) CheckChainID(reference []byte) error {
	if !bytes.Equal(h.ChainID, reference) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			data.ErrInvalidChainID,
			hex.EncodeToString(reference),
			hex.EncodeToString(h.ChainID),
		)
	}

	return nil
}

// Clone the underlying data
func (mb *MiniBlock) Clone() *MiniBlock {
	newMb := &MiniBlock{
		ReceiverShardID: mb.ReceiverShardID,
		SenderShardID:   mb.SenderShardID,
		Type:            mb.Type,
	}
	newMb.TxHashes = make([][]byte, len(mb.TxHashes))
	copy(newMb.TxHashes, mb.TxHashes)

	return newMb
}
