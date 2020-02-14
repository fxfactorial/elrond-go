package shard

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type preProcessorsContainerFactory struct {
	shardCoordinator    sharding.Coordinator
	store               dataRetriever.StorageService
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	dataPool            dataRetriever.PoolsHolder
	addrConverter       state.AddressConverter
	blsAddressConverter state.AddressConverter
	txProcessor         process.TransactionProcessor
	scProcessor         process.SmartContractProcessor
	scResultProcessor   process.SmartContractResultProcessor
	rewardsTxProcessor  process.RewardTransactionProcessor
	accounts            state.AccountsAdapter
	peerAccounts        state.AccountsAdapter
	requestHandler      process.RequestHandler
	rewardsProducer     process.InternalTransactionProducer
	economicsFee        process.FeeHandler
	miniBlocksCompacter process.MiniBlocksCompacter
	gasHandler          process.GasHandler
	blockTracker        preprocess.BlockTracker
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	accounts state.AccountsAdapter,
	peerAccounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	scProcessor process.SmartContractProcessor,
	scResultProcessor process.SmartContractResultProcessor,
	rewardsTxProcessor process.RewardTransactionProcessor,
	rewardsProducer process.InternalTransactionProducer,
	economicsFee process.FeeHandler,
	miniBlocksCompacter process.MiniBlocksCompacter,
	gasHandler process.GasHandler,
	blockTracker preprocess.BlockTracker,
	blsAddressConverter state.AddressConverter,
) (*preProcessorsContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(addrConverter) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(scProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(scResultProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(rewardsTxProcessor) {
		return nil, process.ErrNilRewardsTxProcessor
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(rewardsProducer) {
		return nil, process.ErrNilInternalTransactionProducer
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(miniBlocksCompacter) {
		return nil, process.ErrNilMiniBlocksCompacter
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:    shardCoordinator,
		store:               store,
		marshalizer:         marshalizer,
		hasher:              hasher,
		dataPool:            dataPool,
		addrConverter:       addrConverter,
		txProcessor:         txProcessor,
		accounts:            accounts,
		scProcessor:         scProcessor,
		scResultProcessor:   scResultProcessor,
		rewardsTxProcessor:  rewardsTxProcessor,
		requestHandler:      requestHandler,
		rewardsProducer:     rewardsProducer,
		economicsFee:        economicsFee,
		miniBlocksCompacter: miniBlocksCompacter,
		gasHandler:          gasHandler,
		blockTracker:        blockTracker,
		peerAccounts:        peerAccounts,
		blsAddressConverter: blsAddressConverter,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	preproc, err := ppcm.createTxPreProcessor(block.TxBlock)
	if err != nil {
		return nil, err
	}

	err = container.Add(block.TxBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createTxPreProcessor(block.InvalidBlock)
	if err != nil {
		return nil, err
	}

	err = container.Add(block.InvalidBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createSmartContractResultPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createRewardsTransactionPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.RewardsBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createPeerPreprocessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.PeerBlock, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *preProcessorsContainerFactory) createTxPreProcessor(blockType block.Type) (process.PreProcessor, error) {
	txPreprocessor, err := preprocess.NewTransactionPreprocessor(
		ppcm.dataPool.Transactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.txProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestTransaction,
		ppcm.economicsFee,
		ppcm.miniBlocksCompacter,
		ppcm.gasHandler,
		ppcm.blockTracker,
		blockType,
	)

	return txPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	scrPreprocessor, err := preprocess.NewSmartContractResultPreprocessor(
		ppcm.dataPool.UnsignedTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.scResultProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestUnsignedTransactions,
		ppcm.gasHandler,
		ppcm.economicsFee,
	)

	return scrPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createRewardsTransactionPreProcessor() (process.PreProcessor, error) {
	rewardTxPreprocessor, err := preprocess.NewRewardTxPreprocessor(
		ppcm.dataPool.RewardTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.rewardsTxProcessor,
		ppcm.rewardsProducer,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestRewardTransactions,
		ppcm.gasHandler,
	)

	return rewardTxPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createPeerPreprocessor() (process.PreProcessor, error) {
	txPreprocessor, err := preprocess.NewPeerPreprocessor(
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.peerAccounts,
		ppcm.blsAddressConverter,
		block.PeerBlock,
	)

	return txPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
