package mock

import "github.com/ElrondNetwork/elrond-go/process"

// InterceptorProcessorStub -
type InterceptorProcessorStub struct {
	ValidateCalled func(data process.InterceptedData) error
	SaveCalled     func(data process.InterceptedData) error
}

// Validate -
func (ips *InterceptorProcessorStub) Validate(data process.InterceptedData) error {
	return ips.ValidateCalled(data)
}

// Save -
func (ips *InterceptorProcessorStub) Save(data process.InterceptedData) error {
	return ips.SaveCalled(data)
}

// SignalEndOfProcessing -
func (ips *InterceptorProcessorStub) SignalEndOfProcessing(data []process.InterceptedData) {
}

// IsInterfaceNil -
func (ips *InterceptorProcessorStub) IsInterfaceNil() bool {
	if ips == nil {
		return true
	}
	return false
}
