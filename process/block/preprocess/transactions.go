package preprocess

import (
	"errors"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/block/preprocess")

// TODO: increase code coverage with unit tests

type transactions struct {
	*basePreProcess
	chRcvAllTxs          chan bool
	onRequestTransaction func(shardID uint32, txHashes [][]byte)
	txsForCurrBlock      txsForBlock
	txPool               dataRetriever.ShardedDataCacherNotifier
	storage              dataRetriever.StorageService
	txProcessor          process.TransactionProcessor
	accounts             state.AccountsAdapter
	orderedTxs           map[string][]data.TransactionHandler
	orderedTxHashes      map[string][][]byte
	mutOrderedTxs        sync.RWMutex
	miniBlocksCompacter  process.MiniBlocksCompacter
	blockTracker         BlockTracker
	blockType            block.Type
}

// NewTransactionPreprocessor creates a new transaction preprocessor object
func NewTransactionPreprocessor(
	txDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestTransaction func(shardID uint32, txHashes [][]byte),
	economicsFee process.FeeHandler,
	miniBlocksCompacter process.MiniBlocksCompacter,
	gasHandler process.GasHandler,
	blockTracker BlockTracker,
	blockType block.Type,
) (*transactions, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(txDataPool) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilTxStorage
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestTransaction == nil {
		return nil, process.ErrNilRequestHandler
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

	bpp := basePreProcess{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
		gasHandler:       gasHandler,
		economicsFee:     economicsFee,
	}

	txs := transactions{
		basePreProcess:       &bpp,
		storage:              store,
		txPool:               txDataPool,
		onRequestTransaction: onRequestTransaction,
		txProcessor:          txProcessor,
		accounts:             accounts,
		miniBlocksCompacter:  miniBlocksCompacter,
		blockTracker:         blockTracker,
		blockType:            blockType,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterHandler(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)

	return &txs, nil
}

// waitForTxHashes waits for a call whether all the requested transactions appeared
func (txs *transactions) waitForTxHashes(waitTime time.Duration) error {
	select {
	case <-txs.chRcvAllTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested transactions arrived and were saved into the pool
func (txs *transactions) IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error {
	if requestedTxs > 0 {
		log.Debug("requested missing txs",
			"num txs", requestedTxs)
		err := txs.waitForTxHashes(haveTime())
		txs.txsForCurrBlock.mutTxsForBlock.Lock()
		missingTxs := txs.txsForCurrBlock.missingTxs
		txs.txsForCurrBlock.missingTxs = 0
		txs.txsForCurrBlock.mutTxsForBlock.Unlock()
		log.Debug("received missing txs",
			"num txs", requestedTxs-missingTxs)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveTxBlockFromPools removes transactions and miniblocks from associated pools
func (txs *transactions) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil || body.IsInterfaceNil() {
		return process.ErrNilTxBlockBody
	}
	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return process.ErrNilMiniBlockPool
	}

	err := txs.removeDataFromPools(body, miniBlockPool, txs.txPool, txs.blockType)

	return err
}

// RestoreTxBlockIntoPools restores the transactions and miniblocks to associated pools
func (txs *transactions) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	txsRestored := 0

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		txsBuff, err := txs.storage.GetAll(dataRetriever.TransactionUnit, miniBlock.TxHashes)
		if err != nil {
			log.Debug("tx from mini block was not found in TransactionUnit",
				"sender shard ID", miniBlock.SenderShardID,
				"receiver shard ID", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
			)

			return txsRestored, err
		}

		for txHash, txBuff := range txsBuff {
			tx := transaction.Transaction{}
			err = txs.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return txsRestored, err
			}

			txs.txPool.AddData([]byte(txHash), &tx, strCache)
		}

		miniBlockHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, miniBlock)
		if err != nil {
			return txsRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock)

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(
	body block.Body,
	haveTime func() bool,
) error {

	mapHashesAndTxs := txs.GetAllCurrentUsedTxs()
	expandedMiniBlocks, err := txs.miniBlocksCompacter.Expand(block.MiniBlockSlice(body), mapHashesAndTxs)
	if err != nil {
		return err
	}

	// basic validation already done in interceptors
	for i := 0; i < len(expandedMiniBlocks); i++ {
		miniBlock := expandedMiniBlocks[i]
		if miniBlock.Type != txs.blockType {
			continue
		}

		gasConsumedByMiniBlockInSenderShard := uint64(0)
		gasConsumedByMiniBlockInReceiverShard := uint64(0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			txs.txsForCurrBlock.mutTxsForBlock.RLock()
			txInfoFromMap := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
			txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

			if txInfoFromMap == nil || txInfoFromMap.tx == nil {
				log.Debug("missing transaction in ProcessBlockTransactions ", "type", block.TxBlock, "txHash", txHash)
				return process.ErrMissingTransaction
			}

			tx, ok := txInfoFromMap.tx.(*transaction.Transaction)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err = txs.processAndRemoveBadTransaction(
				txHash,
				tx,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)

			if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
				return err
			}

			err = txs.computeGasConsumed(
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
				tx,
				txHash,
				&gasConsumedByMiniBlockInSenderShard,
				&gasConsumedByMiniBlockInReceiverShard)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SaveTxBlockToStorage saves transactions from body into storage
func (txs *transactions) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		if miniBlock.Type != block.TxBlock {
			continue
		}

		err := txs.saveTxsToStorage(miniBlock.TxHashes, &txs.txsForCurrBlock, txs.storage, dataRetriever.TransactionUnit)
		if err != nil {
			return err
		}
	}

	return nil
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (txs *transactions) receivedTransaction(txHash []byte) {
	receivedAllMissing := txs.baseReceivedTransaction(txHash, &txs.txsForCurrBlock, txs.txPool, txs.blockType)

	if receivedAllMissing {
		txs.chRcvAllTxs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created transactions at this round
func (txs *transactions) CreateBlockStarted() {
	_ = process.EmptyChannel(txs.chRcvAllTxs)

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.missingTxs = 0
	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	txs.mutOrderedTxs.Lock()
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.mutOrderedTxs.Unlock()
}

// RequestBlockTransactions request for transactions if missing from a block.Body
func (txs *transactions) RequestBlockTransactions(body block.Body) int {
	requestedTxs := 0
	missingTxsForShards := txs.computeMissingAndExistingTxsForShards(body)

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for senderShardID, mbsTxHashes := range missingTxsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			txs.setMissingTxsForShard(senderShardID, mbTxHashes)
		}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	for senderShardID, mbsTxHashes := range missingTxsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			requestedTxs += len(mbTxHashes.txHashes)
			txs.onRequestTransaction(senderShardID, mbTxHashes.txHashes)
		}
	}

	return requestedTxs
}

func (txs *transactions) setMissingTxsForShard(senderShardID uint32, mbTxHashes *txsHashesInfo) {
	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: mbTxHashes.receiverShardID}
	for _, txHash := range mbTxHashes.txHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfoToSet}
	}
}

// computeMissingAndExistingTxsForShards calculates what transactions are available and what are missing from block.Body
func (txs *transactions) computeMissingAndExistingTxsForShards(body block.Body) map[uint32][]*txsHashesInfo {
	missingTxsForShard := txs.computeExistingAndMissing(
		body,
		&txs.txsForCurrBlock,
		txs.chRcvAllTxs,
		txs.blockType,
		txs.txPool)

	return missingTxsForShard
}

// processAndRemoveBadTransactions processed transactions, if txs are with error it removes them from pool
func (txs *transactions) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := txs.txProcessor.ProcessTransaction(transaction)
	isTxTargetedForDeletion := err == process.ErrLowerNonceInTransaction || errors.Is(err, process.ErrInsufficientFee)
	if isTxTargetedForDeletion {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(transactionHash, strCache)
	}

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	txShardInfoToSet := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(transactionHash)] = &txInfo{tx: transaction, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return err
}

// RequestTransactionsForMiniBlock requests missing transactions for a certain miniblock
func (txs *transactions) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if miniBlock == nil {
		return 0
	}

	missingTxsForMiniBlock := txs.computeMissingTxsForMiniBlock(miniBlock)
	if len(missingTxsForMiniBlock) > 0 {
		txs.onRequestTransaction(miniBlock.SenderShardID, missingTxsForMiniBlock)
	}

	return len(missingTxsForMiniBlock)
}

// computeMissingTxsForMiniBlock computes missing transactions for a certain miniblock
func (txs *transactions) computeMissingTxsForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	if miniBlock.Type != txs.blockType {
		return nil
	}

	missingTransactions := make([][]byte, 0, len(miniBlock.TxHashes))
	searchFirst := txs.blockType == block.InvalidBlock

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			txs.txPool,
			searchFirst)

		if tx == nil || tx.IsInterfaceNil() {
			missingTransactions = append(missingTransactions, txHash)
		}
	}

	return sliceUtil.TrimSliceSliceByte(missingTransactions)
}

// getAllTxsFromMiniBlock gets all the transactions from a miniblock into a new structure
func (txs *transactions) getAllTxsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*transaction.Transaction, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := txs.txPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilTransactionPool
	}

	// verify if all transaction exists
	txsSlice := make([]*transaction.Transaction, 0, len(mb.TxHashes))
	txHashes := make([][]byte, 0, len(mb.TxHashes))
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilTransaction
		}

		tx, ok := tmp.(*transaction.Transaction)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
		txHashes = append(txHashes, txHash)
		txsSlice = append(txsSlice, tx)
	}

	return txsSlice, txHashes, nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the transactions added into the miniblocks
// as long as it has time
func (txs *transactions) CreateAndProcessMiniBlocks(
	maxTxSpaceRemained uint32,
	maxMbSpaceRemained uint32,
	haveTime func() bool,
) (block.MiniBlockSlice, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	newMBAdded := true
	txSpaceRemained := int(maxTxSpaceRemained)

	miniBlock, err := txs.createAndProcessMiniBlock(
		txs.shardCoordinator.SelfId(),
		sharding.MetachainShardId,
		txSpaceRemained,
		haveTime)

	if err == nil && len(miniBlock.TxHashes) > 0 {
		txSpaceRemained -= len(miniBlock.TxHashes)
		miniBlocks = append(miniBlocks, miniBlock)
	}

	for newMBAdded {
		newMBAdded = false
		for shardId := uint32(0); shardId < txs.shardCoordinator.NumberOfShards(); shardId++ {
			if !haveTime() {
				break
			}

			mbSpaceRemained := int(maxMbSpaceRemained) - len(miniBlocks)
			if mbSpaceRemained <= 0 {
				break
			}

			//TODO: We should analyze if this check could be done more restrictive or not, depending of the pending
			//miniblocks given by the last metablock notarized instead of pending miniblocks given by the last metablock
			//received in block tracker (the current state of all shards)
			if txs.blockTracker.IsShardStuck(shardId) {
				continue
			}

			var miniBlockForShard *block.MiniBlock
			miniBlockForShard, err = txs.createAndProcessMiniBlock(
				txs.shardCoordinator.SelfId(),
				shardId,
				txSpaceRemained,
				haveTime)
			if err != nil {
				continue
			}

			if len(miniBlockForShard.TxHashes) > 0 {
				txSpaceRemained -= len(miniBlockForShard.TxHashes)
				miniBlocks = append(miniBlocks, miniBlockForShard)
				newMBAdded = true
			}
		}
	}

	mapHashesAndTxs := txs.GetAllCurrentUsedTxs()
	compactedMiniBlocks := txs.miniBlocksCompacter.Compact(miniBlocks, mapHashesAndTxs)

	return compactedMiniBlocks, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the transactions added into the miniblock
func (txs *transactions) createAndProcessMiniBlock(
	senderShardId uint32,
	receiverShardId uint32,
	spaceRemained int,
	haveTime func() bool,
) (*block.MiniBlock, error) {
	if txs.blockType != block.TxBlock {
		return &block.MiniBlock{}, nil
	}

	timeBefore := time.Now()
	orderedTxs, orderedTxHashes, err := txs.computeOrderedTxs(senderShardId, receiverShardId)
	timeAfter := time.Now()

	if err != nil {
		log.Trace("computeOrderedTxs", "error", err.Error())
		return nil, err
	}

	if !haveTime() {
		log.Debug("time is up ordering txs",
			"num txs", len(orderedTxs),
			"time [s]", timeAfter.Sub(timeBefore).Seconds(),
		)
		return nil, process.ErrTimeIsOut
	}

	log.Trace("time elapsed to ordered txs",
		"num txs", len(orderedTxs),
		"time [s]", timeAfter.Sub(timeBefore).Seconds(),
	)

	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = senderShardId
	miniBlock.ReceiverShardID = receiverShardId
	miniBlock.TxHashes = make([][]byte, 0)
	miniBlock.Type = block.TxBlock

	addedTxs := 0
	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for index := range orderedTxs {
		txHandler := orderedTxs[index]
		tx := txHandler.(*transaction.Transaction)
		txHash := orderedTxHashes[index]

		if !haveTime() {
			break
		}

		if txs.isTxAlreadyProcessed(txHash, &txs.txsForCurrBlock) {
			continue
		}

		snapshot := txs.accounts.JournalLen()
		oldGasConsumedByMiniBlockInSenderShard := gasConsumedByMiniBlockInSenderShard
		oldGasConsumedByMiniBlockInReceiverShard := gasConsumedByMiniBlockInReceiverShard

		err = txs.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			tx,
			txHash,
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard)

		if err != nil {
			continue
		}

		// execute transaction to change the trie root hash
		err = txs.processAndRemoveBadTransaction(
			txHash,
			tx,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
		)

		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			log.Trace("bad tx",
				"error", err.Error(),
				"hash", txHash,
			)

			err = txs.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Debug("revert to snapshot", "error", err.Error())
			}

			txs.gasHandler.RemoveGasConsumed([][]byte{txHash})
			txs.gasHandler.RemoveGasRefunded([][]byte{txHash})

			gasConsumedByMiniBlockInSenderShard = oldGasConsumedByMiniBlockInSenderShard
			gasConsumedByMiniBlockInReceiverShard = oldGasConsumedByMiniBlockInReceiverShard

			continue
		}

		gasRefunded := txs.gasHandler.GasRefunded(txHash)
		gasConsumedByMiniBlockInReceiverShard -= gasRefunded
		if senderShardId == receiverShardId {
			gasConsumedByMiniBlockInSenderShard -= gasRefunded
		}

		if !errors.Is(err, process.ErrFailedTransaction) {
			miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
		}
		addedTxs++

		if addedTxs >= spaceRemained { // max transactions count in one block was reached
			log.Debug("max txs accepted in one block is reached",
				"num added txs", len(miniBlock.TxHashes),
				"total txs", len(orderedTxs),
			)

			log.Debug("mini block info",
				"gas consumed in sender shard", gasConsumedByMiniBlockInSenderShard,
				"gas consumed in receiver shard", gasConsumedByMiniBlockInReceiverShard,
				"gas consumed in self shard", txs.gasHandler.TotalGasConsumed(),
				"txs ordered", len(orderedTxs),
				"txs added", len(miniBlock.TxHashes))

			return miniBlock, nil
		}
	}

	if addedTxs > 0 {
		log.Debug("mini block info",
			"gas consumed in sender shard", gasConsumedByMiniBlockInSenderShard,
			"gas consumed in receiver shard", gasConsumedByMiniBlockInReceiverShard,
			"gas consumed in self shard", txs.gasHandler.TotalGasConsumed(),
			"txs ordered", len(orderedTxs),
			"txs added", len(miniBlock.TxHashes))
	}

	return miniBlock, nil
}

func (txs *transactions) computeOrderedTxs(
	sndShardId uint32,
	dstShardId uint32,
) ([]data.TransactionHandler, [][]byte, error) {
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txShardPool := txs.txPool.ShardDataStore(strCache)

	if txShardPool == nil {
		return nil, nil, process.ErrNilTxDataPool
	}
	if txShardPool.Len() == 0 {
		return nil, nil, process.ErrEmptyTxDataPool
	}

	sortedTransactionsProvider := createSortedTransactionsProvider(txs, txShardPool, strCache)
	sortedTxs, sortedTxsHashes := sortedTransactionsProvider.GetSortedTransactions()
	return sortedTxs, sortedTxsHashes, nil
}

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
) error {
	if txs.blockType != block.TxBlock {
		return nil
	}

	if miniBlock.Type != block.TxBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	var err error

	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	processedTxHashes := make([][]byte, 0)

	defer func() {
		if err != nil {
			txs.gasHandler.RemoveGasConsumed(processedTxHashes)
			txs.gasHandler.RemoveGasRefunded(processedTxHashes)
		}
	}()

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for index := range miniBlockTxs {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		err = txs.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockTxs[index],
			miniBlockTxHashes[index],
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard)

		if err != nil {
			return err
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	for index := range miniBlockTxs {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index])
		if err != nil {
			return err
		}
	}

	txShardInfoToSet := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockTxs[index], txShardInfo: txShardInfoToSet}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
}

// CreateMarshalizedData marshalizes transactions and creates and saves them into a new structure
func (txs *transactions) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs, err := txs.createMarshalizedData(txHashes, &txs.txsForCurrBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedTxs returns all the transactions used at current creation / processing
func (txs *transactions) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	txPool := make(map[string]data.TransactionHandler, len(txs.txsForCurrBlock.txHashAndInfo))

	txs.txsForCurrBlock.mutTxsForBlock.RLock()
	for txHash, txInfoFromMap := range txs.txsForCurrBlock.txHashAndInfo {
		txPool[txHash] = txInfoFromMap.tx
	}
	txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

	return txPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (txs *transactions) IsInterfaceNil() bool {
	return txs == nil
}
