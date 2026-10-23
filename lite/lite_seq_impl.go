package lite

import "errors"

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
	panic("unimplemented")
}

// Update implements [Sequencer].
func (inst *innerSequencer) Update(ctx *SequencerContext) error {

	if ctx.Updating {
		ctx.Updating = false
	} else {
		return nil
	}

	// do:update

	return nil
}

func (inst *innerSequencer) _impl() Sequencer {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerSequencerTrigger struct {
	context *SequencerContext
	seq     Sequencer
}

// OnClock implements [ClockListener].
func (inst *innerSequencerTrigger) OnClock(ce *ClockEvent) error {

	var results []error
	cc := new(ClockContext)
	tp := ce.Point
	ctx := inst.context

	bank := inst.innerGetCurrentBank()
	bank_body := bank.Body

	stepCount1 := ctx.TotalSteps
	stepCount2 := inst.innerComputeStepCount(tp)
	allSteps := bank_body.Steps.Items()

	defer func() {
		ctx.TotalSteps = stepCount2
	}()

	for count := stepCount1; count < stepCount2; count++ {

		step := inst.innerGetStep(count, allSteps)

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

func (inst *innerSequencerTrigger) innerComputeStepCount(tp *TimePoint) StepCount {

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
	return inst.seq.OnClock(cc)
}

func (inst *innerSequencerTrigger) _impl() ClockListener {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerSequencerFactory struct {
}

// CreateSequencer implements [SequencerFactory].
func (inst *innerSequencerFactory) CreateSequencer(ctx1 *Context) (Sequencer, error) {

	ctx2 := new(SequencerContext)
	seq := new(innerSequencer)
	cfg := ctx1.Configuration
	trigger := new(innerSequencerTrigger)

	seq.context = ctx2

	trigger.context = ctx2
	trigger.seq = seq

	ctx2.Sequencer = seq
	ctx2.Banks.Init(cfg.CountBankMax)
	ctx2.Steps.Init(cfg.CountStepMax)
	ctx2.Tracks.Init(cfg.CountTrackMax)
	ctx1.Clock.AddListener(trigger)

	return ctx2.Sequencer, nil
}

func (inst *innerSequencerFactory) _impl() SequencerFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
