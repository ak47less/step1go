package lite

import (
	"fmt"

	"github.com/ak47less/step1go/common"
)

////////////////////////////////////////////////////////////////////////////////
// innerSequencer

type innerSequencer struct {
	context *SequencerContext
}

// GetBanks implements [Sequencer].
func (inst *innerSequencer) GetBanks() []*Bank {
	return inst.context.Banks.Items()
}

// GetContext implements [Sequencer].
func (inst *innerSequencer) GetContext() *SequencerContext {
	return inst.context
}

// GetSteps implements [Sequencer].
func (inst *innerSequencer) GetSteps() []*Step {
	return inst.context.Steps.Items()
}

// GetTracks implements [Sequencer].
func (inst *innerSequencer) GetTracks() []*Track {
	return inst.context.Tracks.Items()
}

// OnClock implements [Sequencer].
func (inst *innerSequencer) OnStep(stc *StepTickContext) error {

	current := &stc.Current
	loc := current.Index
	idx := current.Location

	logger := common.Log()
	logger.Trace("innerSequencer.OnClock( count=%d, index=%d )", loc, idx)

	return nil
}

// Update implements [Sequencer].
func (inst *innerSequencer) Update(ctx *SequencerContext) error {
	if ctx.Updating {
		up := new(innerSequencerUpdater)
		up.context = inst.context
		return up.update()
	}
	return nil
}

func (inst *innerSequencer) _impl() Sequencer {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerSequencerUpdater struct {
	context *SequencerContext
}

func (inst *innerSequencerUpdater) update() error {

	ctx := inst.context
	logger := common.Log()
	tasks := make([]func(sc *SequencerContext) error, 0)

	ctx.Updating = false
	logger.Info("innerSequencer.Update()")

	tasks = append(tasks, inst.innerDoComputeSpeed)
	tasks = append(tasks, inst.innerDoLoadCurrentBank)

	for _, t := range tasks {
		err := t(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerSequencerUpdater) innerDoComputeSpeed(sc *SequencerContext) error {

	// 计算 sps & spb

	const the60s Seconds = 60 // 60s/min
	const minBpm = 5

	ec := sc.EngineContext
	tl := ec.WorkTileline
	profile := tl.Profile

	bpm := profile.GetSpeed()
	step_p_beat := sc.StepsPerBeat

	if bpm < minBpm {
		return fmt.Errorf("the bpm is too low (bpm=%v)", bpm)
	}

	if step_p_beat < 1 {
		return fmt.Errorf("bad value of steps/beat (value=%d)", step_p_beat)
	}

	sec_p_beat := the60s / Seconds(bpm)

	sc.SecondsPerBeat = sec_p_beat
	sc.SecondsPerStep = sec_p_beat / Seconds(step_p_beat)

	return nil
}

func (inst *innerSequencerUpdater) innerDoLoadCurrentBank(sc *SequencerContext) error {

	banklist := sc.Banks.Items()
	idx := int(sc.CurrentBankIndex)
	size := len(banklist)

	if idx >= size {
		return fmt.Errorf("innerSequencerUpdater: bank index is out of array [index:%d size:%d] ", idx, size)
	}

	bank := banklist[idx]
	if bank == nil {
		return fmt.Errorf("innerSequencerUpdater: bank[%d] is nil", idx)
	}

	// body
	err := inst.innerPrepareBank(bank)
	if err != nil {
		return err
	}

	sc.CurrentBank = bank

	return nil
}

func (inst *innerSequencerUpdater) innerPrepareBank(bank *Bank) error {

	if bank == nil {
		return fmt.Errorf("innerSequencerUpdater: bank is nil")
	}
	sc := bank.Context
	if sc == nil {
		return fmt.Errorf("innerSequencerUpdater: SequencerContext is nil")
	}

	var iBank BankIndex = bank.Index
	var iStep StepIndex = 0
	var iTrack TrackIndex = 0

	stepIdxLimit := StepIndex(sc.Steps.GetSize())
	trackIdxLimit := TrackIndex(sc.Tracks.GetSize())
	want := new(Cell)
	cellman := bank.CellManager
	const autoMake = true

	for iTrack = 0; iTrack < trackIdxLimit; iTrack++ {

		for iStep = 0; iStep < stepIdxLimit; iStep++ {

			want.BankIndex = iBank
			want.StepIndex = iStep
			want.TrackIndex = iTrack

			have, err := cellman.OpenCell(want, autoMake)
			if err != nil {
				return err
			}

			if have.Complex == nil {
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type innerSequencerFactory struct {
}

// CreateSequencer implements [SequencerFactory].
func (inst *innerSequencerFactory) CreateSequencer(ec *Context) (Sequencer, error) {

	sc := new(SequencerContext)
	seq := new(innerSequencer)
	cfg := ec.Configuration
	trigger := new(innerSequencerTrigger)
	cellman := new(CellManager)

	seq.context = sc
	trigger.context = sc

	ec.Clock.AddListener(trigger)

	sc.Sequencer = seq
	sc.Banks.Init(cfg.CountBankMax)
	sc.Steps.Init(cfg.CountStepMax)
	sc.Tracks.Init(cfg.CountTrackMax)
	sc.Engine = ec.Engine
	sc.EngineContext = ec
	sc.Updating = true
	sc.StepsPerBeat = 2
	sc.CellManager = cellman

	return sc.Sequencer, nil
}

func (inst *innerSequencerFactory) _impl() SequencerFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
