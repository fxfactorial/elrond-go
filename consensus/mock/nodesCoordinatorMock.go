package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32) ([]string, error)
}

// ComputeValidatorsGroup -
func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness, round, shardId)
	}

	list := []sharding.Validator{
		NewValidatorMock(big.NewInt(0), 0, []byte("A"), []byte("AA")),
		NewValidatorMock(big.NewInt(0), 0, []byte("B"), []byte("BB")),
		NewValidatorMock(big.NewInt(0), 0, []byte("C"), []byte("CC")),
		NewValidatorMock(big.NewInt(0), 0, []byte("D"), []byte("DD")),
		NewValidatorMock(big.NewInt(0), 0, []byte("E"), []byte("EE")),
		NewValidatorMock(big.NewInt(0), 0, []byte("F"), []byte("FF")),
		NewValidatorMock(big.NewInt(0), 0, []byte("G"), []byte("GG")),
		NewValidatorMock(big.NewInt(0), 0, []byte("H"), []byte("HH")),
		NewValidatorMock(big.NewInt(0), 0, []byte("I"), []byte("II")),
	}

	return list, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(uint32) int {
	return 1
}

// GetAllValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllValidatorsPublicKeys() map[uint32][][]byte {
	return nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(publicKeys []string) []uint64 {
	return nil
}

// GetValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetValidatorsRewardsAddresses -
func (ncm *NodesCoordinatorMock) GetValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsRewardsAddressesCalled(randomness, round, shardId)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0)
	for _, v := range validators {
		addresses = append(addresses, string(v.Address()))
	}

	return addresses, nil
}

// SetNodesPerShards -
func (ncm *NodesCoordinatorMock) SetNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

// SetConsensusGroupSize -
func (ncm *NodesCoordinatorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32) (publicKeys []string, err error) {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	panic("implement me")
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	if ncm == nil {
		return true
	}
	return false
}
