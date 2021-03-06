package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestConsensusCore_WithNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		nil,
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestConsensusCore_WithNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		nil,
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestConsensusCore_WithNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		nil,
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestConsensusCore_WithNilBroadcastMessengerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		nil,
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBroadcastMessenger, err)
}

func TestConsensusCore_WithNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		nil,
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestConsensusCore_WithNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		nil,
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestConsensusCore_WithNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		nil,
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestConsensusCore_WithNilBlsPrivateKeyShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		nil,
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsPrivateKey, err)
}

func TestConsensusCore_WithNilBlsSingleSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		nil,
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsSingleSigner, err)
}

func TestConsensusCore_WithNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		nil,
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestConsensusCore_WithNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		nil,
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestConsensusCore_WithNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		nil,
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestConsensusCore_WithNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		nil,
		consensusCoreMock.SyncTimer(),
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilValidatorGroupSelector, err)
}

func TestConsensusCore_WithNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		nil,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestConsensusCore_CreateConsensusCoreShouldWork(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.BroadcastMessenger(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.PrivateKey(),
		consensusCoreMock.SingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.NodesCoordinator(),
		consensusCoreMock.SyncTimer(),
	)

	assert.NotNil(t, consensusCore)
	assert.Nil(t, err)
}
