package requestHandlers

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeoutSendRequests = time.Second * 2

func createResolversFinderStubThatShouldNotBeCalled(tb testing.TB) *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "IntraShardResolverCalled should not have been called")
			return nil, nil
		},
		MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "MetaChainResolverCalled should not have been called")
			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "CrossShardResolverCalled should not have been called")
			return nil, nil
		},
	}
}

//------- NewMetaResolver

func TestNewMetaResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		nil,
		&mock.RequestedItemsHandlerStub{},
		100,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewMetaResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		nil,

		100,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilRequestedItemsHandler, err)
}

func TestNewMetaResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},

		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewMetaResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},

		100,
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(rrh))
}

//------- NewShardResolver

func TestNewShardResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		nil,
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewShardResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		nil,

		1,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilRequestedItemsHandler, err)
}

func TestNewShardResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},

		0,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewShardResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

//------- RequestTransaction

func TestResolverRequestHandler_RequestTransactionErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestTransaction(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestTransactionWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	wrongTxResolver := &mock.HeaderResolverStub{}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestTransaction(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestTransactionShouldRequestTransactions(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestTransactionErrorsOnRequestShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

//------- RequestMiniBlock

func TestResolverRequestHandler_RequestMiniBlockErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestMiniBlock(0, make([]byte, 0))
}

func TestResolverRequestHandler_RequestMiniBlockErrorsOnRequestShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))

	assert.True(t, wasCalled)
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallWithTheCorrectEpoch(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(7)
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			assert.Equal(t, expectedEpoch, epoch)
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.SetEpoch(expectedEpoch)

	rrh.RequestMiniBlock(0, []byte("mbHash"))
}

//------- RequestShardHeader

func TestResolverRequestHandler_RequestShardHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		1,
		0,
	)

	rrh.RequestShardHeader(0, make([]byte, 0))
}

func TestResolverRequestHandler_RequestShardHeaderHashBadRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeader(1, make([]byte, 0))
}

func TestResolverRequestHandler_RequestShardHeaderShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeader(0, []byte("hdrHash"))

	assert.True(t, wasCalled)
}

//------- RequestMetaHeader

func TestResolverRequestHandler_RequestMetadHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},

		1,
		0,
	)

	rrh.RequestMetaHeader(make([]byte, 0))
}

func TestResolverRequestHandler_RequestMetadHeaderHashNotHeaderResolverShouldNotRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestMetaHeader([]byte("hdrHash"))

	assert.False(t, wasCalled)
}

func TestResolverRequestHandler_RequestMetaHeaderShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestMetaHeader([]byte("hdrHash"))

	assert.True(t, wasCalled)
}

//------- RequestShardHeaderByNonce

func TestResolverRequestHandler_RequestShardHeaderByNonceAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	called := false
	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				called = true
				return true
			},
		},

		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)
	require.True(t, called)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceBadRequest(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	called := false
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
				called = true
				return nil, localErr
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		sharding.MetachainShardId,
	)

	rrh.RequestShardHeaderByNonce(1, 0)
	require.True(t, called)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceFinderReturnsErrorShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceFinderReturnsAWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	hdrResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceResolverFailsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)

	assert.True(t, wasCalled)
}

//------- RequestMetaHeaderByNonce

func TestResolverRequestHandler_RequestMetaHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},

		1,
		0,
	)

	rrh.RequestMetaHeaderByNonce(0)
}

func TestResolverRequestHandler_RequestMetaHeaderByNonceShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		100,
		0,
	)

	rrh.RequestMetaHeaderByNonce(0)

	assert.True(t, wasCalled)
}

//------- RequestSmartContractResult

func TestResolverRequestHandler_RequestScrErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestUnsignedTransactions(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestScrWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	wrongTxResolver := &mock.HeaderResolverStub{}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestUnsignedTransactions(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestScrShouldRequestScr(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestUnsignedTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestScrErrorsOnRequestShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestUnsignedTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

//------- RequestRewardTransaction

func TestResolverRequestHandler_RequestRewardShouldRequestReward(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestRewardTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestRequestTrieNodes_ShouldWork(t *testing.T) {
	t.Parallel()

	called := false
	resolverMock := &mock.HashSliceResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			called = true
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return resolverMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.RequestTrieNodes(0, []byte("hash"), "topic")
	assert.True(t, called)
}

func TestRequestTrieNodes_NilResolver(t *testing.T) {
	t.Parallel()

	localError := errors.New("test error")
	called := false
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				called = true
				return nil, localError
			},
		},
		&mock.RequestedItemsHandlerStub{},

		1,
		0,
	)

	rrh.RequestTrieNodes(sharding.MetachainShardId, []byte("hash"), "topic")
	assert.True(t, called)
}

func TestRequestTrieNodes_RequestByHashError(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
	resolverMock := &mock.HashSliceResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			called = true
			return localError
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return resolverMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.RequestTrieNodes(0, []byte("hash"), "topic")
	assert.True(t, called)
}

func TestRequestStartOfEpochMetaBlock_MissingResolver(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				called = true
				return nil, localError
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
}

func TestRequestStartOfEpochMetaBlock_WrongResolver(t *testing.T) {
	t.Parallel()

	called := false
	resolverMock := &mock.HashSliceResolverStub{}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				called = true
				return resolverMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
}

func TestRequestStartOfEpochMetaBlock_RequestDataFromEpochError(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
	resolverMock := &mock.HeaderResolverStub{
		RequestDataFromEpochCalled: func(identifier []byte) error {
			called = true
			return localError
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				return resolverMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		1,
		0,
	)

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
}

func TestRequestStartOfEpochMetaBlock_AddError(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
	resolverMock := &mock.HeaderResolverStub{
		RequestDataFromEpochCalled: func(identifier []byte) error {
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				return resolverMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{
			AddCalled: func(key string) error {
				called = true
				return localError
			},
		},
		1,
		0,
	)

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
}
