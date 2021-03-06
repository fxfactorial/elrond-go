package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/data/state"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const emptyExcludePeersOnTopic = ""

type resolversContainerFactory struct {
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	dataPools                dataRetriever.PoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	intRandomizer            dataRetriever.IntRandomizer
	dataPacker               dataRetriever.DataPacker
	triesContainer           state.TriesHolder
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.PoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	dataPacker dataRetriever.DataPacker,
	triesContainer state.TriesHolder,
	sizeCheckDelta uint32,
) (*resolversContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(store) {
		return nil, dataRetriever.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}
	if check.IfNil(dataPools) {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(dataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(triesContainer) {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               dataPacker,
		triesContainer:           triesContainer,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, interceptorSlice, err := rcf.generateShardHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	metaKeys, metaInterceptorSlice, err := rcf.generateMetaChainHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(metaKeys, metaInterceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err := rcf.generateTxResolvers(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
		rcf.dataPools.Transactions(),
	)
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
		rcf.dataPools.UnsignedTransactions(),
	)
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateTrieNodesResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver dataRetriever.Resolver,
	createChannel bool,
) (dataRetriever.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

//------- Shard header resolvers

func (rcf *resolversContainerFactory) generateShardHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := emptyExcludePeersOnTopic

		resolver, err := rcf.createShardHeaderResolver(identifierHeader, excludePeersFromTopic, idx)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierHeader
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createShardHeaderResolver(topic string, excludedTopic string, shardID uint32) (dataRetriever.Resolver, error) {
	hdrStorer := rcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		topic,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		shardID,
	)
	if err != nil {
		return nil, err
	}

	//TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore := rcf.store.GetStorer(hdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Meta header resolvers

func (rcf *resolversContainerFactory) generateMetaChainHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	identifierHeader := factory.MetachainBlocksTopic
	resolver, err := rcf.createMetaChainHeaderResolver(identifierHeader, sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHeader}, []dataRetriever.Resolver{resolver}, nil
}

func (rcf *resolversContainerFactory) createMetaChainHeaderResolver(identifier string, shardId uint32) (dataRetriever.Resolver, error) {
	hdrStorer := rcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, identifier, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		shardId,
	)
	if err != nil {
		return nil, err
	}

	hdrNonceStore := rcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Tx resolvers

func (rcf *resolversContainerFactory) generateTxResolvers(
	topic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) ([]string, []dataRetriever.Resolver, error) {

	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

		resolver, err := rcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := rcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierTx

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createTxResolver(
	topic string,
	excludedTopic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) (dataRetriever.Resolver, error) {

	txStorer := rcf.store.GetStorer(unit)

	resolverSender, err := rcf.createOneResolverSender(topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewTxResolver(
		resolverSender,
		dataPool,
		txStorer,
		rcf.marshalizer,
		rcf.dataPacker,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- MiniBlocks resolvers

func (rcf *resolversContainerFactory) generateMiniBlocksResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

		resolver, err := rcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := rcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createMiniBlocksResolver(topic string, excludedTopic string) (dataRetriever.Resolver, error) {
	miniBlocksStorer := rcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	resolverSender, err := rcf.createOneResolverSender(topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPools.MiniBlocks(),
		miniBlocksStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		txBlkResolver,
		false)
}

func (rcf *resolversContainerFactory) createOneResolverSender(
	topic string,
	excludedTopic string,
) (dataRetriever.TopicResolverSender, error) {

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	//TODO instantiate topic sender resolver with the shard IDs for which this resolver is supposed to serve the data
	// this will improve the serving of transactions as the searching will be done only on 2 sharded data units
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		topic,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		uint32(0),
	)
	if err != nil {
		return nil, err
	}

	return resolverSender, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}

func (rcf *resolversContainerFactory) generateTrieNodesResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(i)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice = append(resolverSlice, resolver)
		keys = append(keys, identifierTrieNodes)

		identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(i)
		resolver, err = rcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice = append(resolverSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err = rcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(topic string, trieId string) (dataRetriever.Resolver, error) {
	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		topic,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		rcf.shardCoordinator.SelfId(),
	)
	if err != nil {
		return nil, err
	}

	trie := rcf.triesContainer.Get([]byte(trieId))
	resolver, err := resolvers.NewTrieNodeResolver(
		resolverSender,
		trie,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}
