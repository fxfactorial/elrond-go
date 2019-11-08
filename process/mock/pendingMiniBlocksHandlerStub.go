package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type PendingMiniBlocksHandlerStub struct {
	PendingMiniBlockHeadersCalled func() []block.ShardMiniBlockHeader
	AddProcessedHeaderCalled      func(handler data.HeaderHandler) error
	RevertHeaderCalled            func(handler data.HeaderHandler) error
}

func (p *PendingMiniBlocksHandlerStub) PendingMiniBlockHeaders() []block.ShardMiniBlockHeader {
	if p.PendingMiniBlockHeadersCalled != nil {
		return p.PendingMiniBlockHeadersCalled()
	}
	return nil
}

func (p *PendingMiniBlocksHandlerStub) AddProcessedHeader(handler data.HeaderHandler) error {
	if p.AddProcessedHeaderCalled != nil {
		return p.AddProcessedHeaderCalled(handler)
	}
	return nil
}

func (p *PendingMiniBlocksHandlerStub) RevertHeader(handler data.HeaderHandler) error {
	if p.RevertHeaderCalled != nil {
		return p.RevertHeaderCalled(handler)
	}
	return nil
}

func (p *PendingMiniBlocksHandlerStub) IsInterfaceNil() bool {
	return p == nil
}