package track

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgBaseTracker holds all dependencies required by the process data factory in order to create
// new instances of shard/meta block tracker
type ArgBaseTracker struct {
	Hasher           hashing.Hasher
	HeaderValidator  process.HeaderConstructionValidator
	Marshalizer      marshal.Marshalizer
	RequestHandler   process.RequestHandler
	Rounder          consensus.Rounder
	ShardCoordinator sharding.Coordinator
	Store            dataRetriever.StorageService
	StartHeaders     map[uint32]data.HeaderHandler
}

// ArgShardTracker holds all dependencies required by the process data factory in order to create
// new instances of shard block tracker
type ArgShardTracker struct {
	ArgBaseTracker
	PoolsHolder dataRetriever.PoolsHolder
}

// ArgMetaTracker holds all dependencies required by the process data factory in order to create
// new instances of meta block tracker
type ArgMetaTracker struct {
	ArgBaseTracker
	PoolsHolder dataRetriever.PoolsHolder
}
