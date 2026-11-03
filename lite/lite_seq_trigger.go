package lite

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// StepTickPos

type StepTickPos struct {
	Position RelativePosition // 时间轴位置: at work-timeline
	Index    StepIndex        // 物理下标: index for list of [] Step
	Location StepLocation     // 逻辑下标: @work.timeline
}

////////////////////////////////////////////////////////////////////////////////
// Step_Tick_Clock

type StepTickContext struct {

	// CurrentStep     *Step // 这里不提供具体的step, 以便适配不同的bank

	context *SequencerContext

	wktline *Timeline // the work-timeline

	StepLength   Seconds // 每一步的时长
	StepCount    uint    // 每一个 LOOP 包含的步数
	PreloadSteps uint    // 需要提前加载的步数
	BreakCount   uint    // 中断计数器
	Broken       bool

	//  prev|current|head||realtime

	Ahead    StepTickPos
	Current  StepTickPos
	Realtime StepTickPos
	Previous StepTickPos

	// list []*StepTickPos
}

////////////////////////////////////////////////////////////////////////////////
// StepTickHandler

type StepTickHandler interface {
	OnStep(cc *StepTickContext) error
}

////////////////////////////////////////////////////////////////////////////////
// innerSequencerTrigger

type innerSequencerTrigger struct {
	context *SequencerContext
	// stc     *StepTickContext
}

// OnClock implements [ClockListener].
func (inst *innerSequencerTrigger) OnClock(ce *ClockEvent) error {

	var err error
	stc := inst.context.STC

	// do: prepare stc
	stc, err = inst.innerDoPrepare(stc)
	if err != nil {
		return err
	}

	// do: check-params
	err = inst.innerDoCheckParams(stc)
	if err != nil {
		return err
	}

	// do: compute real-time (fine)
	err = inst.innerDoComputeRealtimePos(stc, ce)
	if err != nil {
		return err
	}

	// do: compute current (align)
	err = inst.innerDoComputeCurrentPos(stc)
	if err != nil {
		return err
	}

	// do: compute ahead (preload)
	err = inst.innerDoComputeAheadPos(stc)
	if err != nil {
		return err
	}

	// do: fast-forward (move prev)
	err = inst.innerDoFastForward(stc)
	if err != nil {
		return err
	}

	inst.context.STC = stc
	return nil
}

func (inst *innerSequencerTrigger) innerNormalizeSTP(stp *StepTickPos, stc *StepTickContext, align bool) error {

	const (
		min_count   = 4
		min_pos     = 0.001
		min_steplen = 0.001
	)

	pos := stp.Position.Position
	steplen := stc.StepLength
	count := stc.StepCount

	if count < min_count {
		count = min_count
	}

	if pos < min_pos {
		pos = 0
	}

	if steplen < min_steplen {
		steplen = min_steplen
	}

	// for location
	location := StepLocation(pos / steplen)
	stp.Location = location

	// for index
	idx := (uint(location) % count)
	stp.Index = StepIndex(idx)

	// align
	if align {
		pos = (steplen * Seconds(location))
		stp.Position.Position = pos
	}

	return nil
}

func (inst *innerSequencerTrigger) innerDoPrepare(stc *StepTickContext) (*StepTickContext, error) {

	if stc == nil {
		stc = inst.context.STC
	}

	if stc == nil {
		stc = new(StepTickContext)
	}

	seqctx := inst.context
	engctx := seqctx.EngineContext

	stc.wktline = engctx.WorkTileline
	stc.context = seqctx

	count := seqctx.Steps.GetSize()
	preload := count / 4

	stc.StepCount = count
	stc.StepLength = seqctx.SecondsPerStep
	stc.PreloadSteps = preload

	return stc, nil
}

func (inst *innerSequencerTrigger) innerDoCheckParams(stc *StepTickContext) error {

	ctx := stc.context
	tline := stc.wktline
	count := stc.StepCount
	steplen := stc.StepLength
	preload := stc.PreloadSteps

	const (
		minStepLen = 0.001
		minPreload = 1
		minCount   = 4
	)

	if ctx == nil {
		return fmt.Errorf("stc.context is nil")
	}
	if tline == nil {
		return fmt.Errorf("stc.wktlin is nil")
	}
	if count < minCount {
		return fmt.Errorf("bad stc.StepCount value: %v", count)
	}
	if preload < minPreload {
		return fmt.Errorf("bad stc.Preload value: %v", preload)
	}
	if steplen < minStepLen {
		return fmt.Errorf("bad stc.StepLength value: %v", steplen)
	}

	stc.Ahead.Position.At = tline
	stc.Current.Position.At = tline
	stc.Previous.Position.At = tline
	stc.Realtime.Position.At = tline

	return nil
}

func (inst *innerSequencerTrigger) innerDoComputeRealtimePos(stc *StepTickContext, ce *ClockEvent) error {

	t1 := ce.Time
	wktlin := stc.wktline
	dst := &stc.Realtime
	abs := t1.Absolute()

	abs.Relative(wktlin, &dst.Position)

	const align = false
	return inst.innerNormalizeSTP(dst, stc, align)
}

func (inst *innerSequencerTrigger) innerDoComputeCurrentPos(stc *StepTickContext) error {

	src := &stc.Realtime // src
	dst := &stc.Current  // dst

	dst.Position = src.Position

	const align = true
	return inst.innerNormalizeSTP(dst, stc, align)
}

func (inst *innerSequencerTrigger) innerDoComputeAheadPos(stc *StepTickContext) error {

	ahead := &stc.Ahead     // dst
	current := &stc.Current // src

	sps := stc.StepLength       // in (sec/step)
	preload := stc.PreloadSteps // in (step)

	t1 := current.Position
	t1.Position += (sps * Seconds(preload))
	ahead.Position = t1

	const align = true
	return inst.innerNormalizeSTP(ahead, stc, align)
}

func (inst *innerSequencerTrigger) innerDoFastForward(stc *StepTickContext) error {

	from := &stc.Previous
	to := &stc.Ahead
	steplen := stc.StepLength
	count := stc.StepCount

	var current StepTickPos
	current = *from

	for current.Location < to.Location {

		gap := to.Location - current.Location
		if uint(gap) > (count / 2) {
			stc.BreakCount++
			stc.Broken = true
			break
		}

		// dispatch clock
		stc.Current = current
		err := inst.innerDispatchOnTickEvent(stc)
		handleError(err)

		// goto next
		current.Position.Position += steplen
		current.Location++
		current.Index = StepIndex(uint(current.Location) % count)
	}

	current = *to
	stc.Previous = current
	return nil
}

func (inst *innerSequencerTrigger) innerDispatchOnTickEvent(stc *StepTickContext) error {
	seq := stc.context.Sequencer
	return seq.OnStep(stc)
}

func (inst *innerSequencerTrigger) _impl() ClockListener {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
