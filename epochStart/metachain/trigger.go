package metachain

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// ArgsNewMetaEpochStartTrigger defines struct needed to create a new start of epoch trigger
type ArgsNewMetaEpochStartTrigger struct {
	GenesisTime     time.Time
	Settings        *config.EpochStartConfig
	Epoch           uint32
	EpochStartRound uint64
}

type trigger struct {
	isEpochStart           bool
	epoch                  uint32
	currentRound           uint64
	currEpochStartRound    uint64
	prevEpochStartRound    uint64
	roundsPerEpoch         uint64
	minRoundsBetweenEpochs uint64
	epochStartMetaHash     []byte
	epochStartTime         time.Time
	mutTrigger             sync.RWMutex
}

// NewEpochStartTrigger creates a trigger for start of epoch
func NewEpochStartTrigger(args *ArgsNewMetaEpochStartTrigger) (*trigger, error) {
	if args == nil {
		return nil, epochStart.ErrNilArgsNewMetaEpochStartTrigger
	}
	if args.Settings == nil {
		return nil, epochStart.ErrNilEpochStartSettings
	}
	if args.Settings.RoundsPerEpoch < 1 {
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs < 1 {
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs > args.Settings.RoundsPerEpoch {
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
	}

	return &trigger{
		roundsPerEpoch:         uint64(args.Settings.RoundsPerEpoch),
		epochStartTime:         args.GenesisTime,
		currEpochStartRound:    args.EpochStartRound,
		prevEpochStartRound:    args.EpochStartRound,
		epoch:                  args.Epoch,
		minRoundsBetweenEpochs: uint64(args.Settings.MinRoundsBetweenEpochs),
		mutTrigger:             sync.RWMutex{},
	}, nil
}

// IsEpochStart return true if conditions are fulfilled for start of epoch
func (t *trigger) IsEpochStart() bool {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.isEpochStart
}

// EpochStartRound returns the start round of the current epoch
func (t *trigger) EpochStartRound() uint64 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.currEpochStartRound
}

// ForceEpochStart sets the conditions for start of epoch to true in case of edge cases
func (t *trigger) ForceEpochStart(round uint64) error {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.currentRound > round {
		return epochStart.ErrSavedRoundIsHigherThanInput
	}
	if t.currentRound == round {
		return epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound
	}

	t.currentRound = round

	if t.currentRound-t.currEpochStartRound < t.minRoundsBetweenEpochs {
		return epochStart.ErrNotEnoughRoundsBetweenEpochs
	}

	t.epoch += 1
	t.currEpochStartRound = t.currentRound
	t.isEpochStart = true

	return nil
}

// Update processes changes in the trigger
func (t *trigger) Update(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.currentRound+1 != round {
		return
	}

	t.currentRound = round

	if t.currentRound > t.currEpochStartRound+t.roundsPerEpoch {
		t.prevEpochStartRound = t.currEpochStartRound
		t.epoch += 1
		t.isEpochStart = true
		t.currEpochStartRound = t.currentRound
	}
}

// Processed sets start of epoch to false and cleans underlying structure
func (t *trigger) Processed() {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.isEpochStart = false
}

// Revert sets the start of epoch back to true
func (t *trigger) Revert() {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.isEpochStart = true
	t.currEpochStartRound = t.prevEpochStartRound
	t.epoch -= 1
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// ReceivedHeader saved the header into pool to verify if end-of-epoch conditions are fulfilled
func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	return t.epochStartMetaHash
}

// SetEpochStartMetaHdrHash sets the epoch start meta header hase
func (t *trigger) SetEpochStartMetaHdrHash(metaHdrHash []byte) {
	t.epochStartMetaHash = metaHdrHash
}

// IsInterfaceNil return true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}