package bootstrapStorage_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapStorer_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	bt, err := bootstrapStorage.NewBootstrapStorer(marshalizer, nil)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilBootStorer, err)
}

func TestNewBootstrapStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	bt, err := bootstrapStorage.NewBootstrapStorer(nil, storer)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilMarshalizer, err)
}

func TestNewBootstrapStorer_ShouldWork(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerMock{}
	marshalizer := &mock.MarshalizerMock{}
	bt, err := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	assert.NotNil(t, bt)
	assert.Nil(t, err)
	assert.False(t, bt.IsInterfaceNil())
}

func TestBootstrapStorer_PutAndGet(t *testing.T) {
	t.Parallel()

	numRounds := int64(10)
	round := int64(0)
	storer := mock.NewStorerMock()
	marshalizer := &mock.MarshalizerMock{}
	bt, _ := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{ShardId: 2, Nonce: 3, Hash: []byte("Hash")}
	dataBoot := bootstrapStorage.BootstrapData{
		LastHeader:                headerInfo,
		LastCrossNotarizedHeaders: []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
		LastSelfNotarizedHeaders:  []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
	}

	err := bt.Put(round, dataBoot)
	assert.Nil(t, err)

	for i := int64(0); i < numRounds; i++ {
		round = i
		err = bt.Put(round, dataBoot)
		assert.Nil(t, err)
	}

	round = bt.GetHighestRound()
	for i := numRounds - 1; i >= 0; i-- {
		dataBoot.LastRound = i - 1
		if i == 0 {
			dataBoot.LastRound = 0
		}
		data, err := bt.Get(round)
		assert.Nil(t, err)
		assert.Equal(t, dataBoot, data)
		round--
	}
}

func TestBootstrapStorer_SaveLastRound(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	roundInStorage := int64(5)
	storer := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			putWasCalled = true
			err := json.Unmarshal(data, &roundInStorage)
			if err != nil {
				fmt.Println(err.Error())
			}
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			k := []byte(strconv.FormatInt(roundInStorage, 10))
			return k, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	bt, _ := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	assert.Equal(t, roundInStorage, bt.GetHighestRound())
	newRound := int64(37)
	err := bt.SaveLastRound(newRound)
	assert.Equal(t, newRound, bt.GetHighestRound())
	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}

func TestTrimHeaderInfoSlice(t *testing.T) {
	t.Parallel()

	input := make([]bootstrapStorage.BootstrapHeaderInfo, 0, 5)
	input = append(input, bootstrapStorage.BootstrapHeaderInfo{})
	input = append(input, bootstrapStorage.BootstrapHeaderInfo{})

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 5, cap(input))

	input = bootstrapStorage.TrimHeaderInfoSlice(input)

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))
}
