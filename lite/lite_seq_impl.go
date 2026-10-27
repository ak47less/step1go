package lite

import (
	"errors"
	"fmt"

	"github.com/ak47less/step1go/common"
)

////////////////////////////////////////////////////////////////////////////////

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
func (inst *innerSequencer) OnClock(cc *ClockContext) error {

	idx := cc.CurrentStepIndex
	cnt := cc.CurrentStepCount

	logger := common.Log()
	logger.Info("innerSequencer.OnClock( count=%d, index=%d )", cnt, idx)

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

type innerSequencerTrigger struct {
	context *SequencerContext
	// seq     Sequencer
}

// OnClock implements [ClockListener].
func (inst *innerSequencerTrigger) OnClock(ce *ClockEvent) error {

	var results []error
	cc := new(ClockContext)
	tp := ce.Time
	ctx := inst.context

	bank := inst.innerGetCurrentBank()
	if bank == nil {
		return fmt.Errorf("innerSequencerTrigger: current bank is nil")
	}
	bank_body := bank.Body

	stepCount1 := ctx.TotalSteps
	stepCount2 := inst.innerComputeStepCount(tp)
	allSteps := bank_body.Steps.Items()

	defer func() {
		ctx.TotalSteps = stepCount2
	}()

	for count := stepCount1; count < stepCount2; count++ {

		step := inst.innerGetStep(count, allSteps)

		if step == nil {
			continue
		}

		cc.CurrentStepIndex = step.Head.Index
		cc.CurrentStepHead = step.Head
		cc.CurrentStepBody = step.Body

		err := inst.innerOnStep(cc)
		if err != nil {
			results = append(results, err)
		}
	}

	if results == nil {
		return nil
	}
	return errors.Join(results...)
}

func (inst *innerSequencerTrigger) innerGetCurrentBank() *Bank {
	ctx := inst.context
	return ctx.CurrentBank
}

func (inst *innerSequencerTrigger) innerGetStep(count StepCount, all []*Step) *Step {
	size := len(all)
	total := int(count)
	if size < 1 {
		return nil
	}
	idx := (total % size)
	return all[idx]
}

func (inst *innerSequencerTrigger) innerComputeStepCount(tp *RelativePosition) StepCount {

	const min = 0.001
	pos := tp.Position
	sps := inst.context.SecondsPerStep

	if sps < min {
		return 0
	}

	count := (pos / sps)
	return StepCount(count)
}

func (inst *innerSequencerTrigger) innerOnStep(cc *ClockContext) error {
	return inst.context.Sequencer.OnClock(cc)
}

func (inst *innerSequencerTrigger) _impl() ClockListener {
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
	bpm := tl.BPM
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
	err := inst.innerPrepareBank(sc, bank)
	if err != nil {
		return err
	}

	sc.CurrentBank = bank

	return nil
}

// func (inst *innerSequencerUpdater) innerPrepareCell(sc *SequencerContext, cell *Cell) error {}
// func (inst *innerSequencerUpdater) innerPrepareStep(sc *SequencerContext, step *Step) error {}

func (inst *innerSequencerUpdater) innerPrepareBank(sc *SequencerContext, bank *Bank) error {

	if sc == nil {
		return fmt.Errorf("innerSequencerUpdater: SequencerContext is nil")
	}
	if bank == nil {
		return fmt.Errorf("innerSequencerUpdater: bank is nil")
	}

	bankbody := bank.Body
	if bankbody == nil {
		bankbody = new(BankBody)
		bankbody.Head = bank.Head
		bank.Body = bankbody
	}

	var iBank BankIndex = bank.Head.Index
	var iStep StepIndex = 0
	var iTrack TrackIndex = 0

	stepIdxLimit := StepIndex(sc.Steps.GetSize())
	trackIdxLimit := TrackIndex(sc.Tracks.GetSize())
	cell := new(Cell)
	cellman := &bankbody.CellManager
	const autoMake = true

	for iTrack = 0; iTrack < trackIdxLimit; iTrack++ {

		for iStep = 0; iStep < stepIdxLimit; iStep++ {

			cell.BankIndex = iBank
			cell.StepIndex = iStep
			cell.TrackIndex = iTrack

			have, err := cellman.OpenCell(cell, autoMake)
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

	return sc.Sequencer, nil
}

func (inst *innerSequencerFactory) _impl() SequencerFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
