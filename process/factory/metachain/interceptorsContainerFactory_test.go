package metachain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

const maxTxNonceDeltaAllowed = 100

var chainID = []byte("chain ID")
var errExpected = errors.New("expected error")

func createStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
	return &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			if matchStrToErrOnCreate == "" {
				return nil
			}

			if strings.Contains(name, matchStrToErrOnCreate) {
				return errExpected
			}

			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			if matchStrToErrOnRegister == "" {
				return nil
			}

			if strings.Contains(topic, matchStrToErrOnRegister) {
				return errExpected
			}

			return nil
		},
	}
}

func createDataPools() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		TrieNodesCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}

	return pools
}

func createStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory

func TestNewInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		nil,
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsContainerFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		nil,
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		nil,
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		nil,
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		1,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsContainerFactory_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		nil,
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewInterceptorsContainerFactory_NilAddrConvShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		nil,
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		nil,
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptorsContainerFactory_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		nil,
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewInterceptorsContainerFactory_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		nil,
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlackListHandler, err)
}

func TestNewInterceptorsContainerFactory_EmptyCahinIDShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		nil,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewInterceptorsContainerFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		nil,
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestNewInterceptorsContainerFactory_EpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		nil,
		0,
		&mock.ValidityAttesterStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}
func TestNewInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

func TestNewInterceptorsContainerFactory_ShouldWorkWithSizeCheck(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		1,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
	assert.False(t, icf.IsInterfaceNil())
}

//------- Create

func TestInterceptorsContainerFactory_CreateTopicMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createStubTopicHandler(factory.MetachainBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createStubTopicHandler(factory.ShardBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterForMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createStubTopicHandler("", factory.MetachainBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createStubTopicHandler("", factory.ShardBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createStubTopicHandler("", factory.AccountTrieNodesTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	nodesCoordinator := &mock.NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		NbShards:           uint32(noOfShards),
		ShardId:            1,
	}

	icf, _ := metachain.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	numInterceptorsMetablock := 1
	numInterceptorsShardHeadersForMetachain := noOfShards
	numInterceptorsTransactionsForMetachain := noOfShards + 1
	numInterceptorsMiniBlocksForMetachain := noOfShards + 1
	numInterceptorsUnsignedTxsForMetachain := noOfShards
	numInterceptorsTrieNodes := (noOfShards + 1) * 2
	totalInterceptors := numInterceptorsMetablock + numInterceptorsShardHeadersForMetachain + numInterceptorsTrieNodes +
		numInterceptorsTransactionsForMetachain + numInterceptorsUnsignedTxsForMetachain + numInterceptorsMiniBlocksForMetachain

	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors, container.Len())
}
