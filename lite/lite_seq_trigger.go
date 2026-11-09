package lite

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// StepTickPos

type StepTickPos struct {
	Position RelativePosition // 时间轴位置: at work-timeline
	Index    StepIndex        // 物理下标: index for list of [] Step
	Location StepLocation     // 逻辑下标: @work.timeline
}

// 根据 location 计算 index
func (inst *StepTickPos) ComputeIndex(info *SequencerInfo, rewrite bool) StepIndex {

	if inst == nil || info == nil {
		return 0
	}

	spc := info.StepsPerCycle
	loc := uint(inst.Location)

	if spc < 1 {
		spc = 1
	}

	idx := StepIndex(loc % spc)
	if rewrite {
		inst.Index = idx
	}
	return idx
}

////////////////////////////////////////////////////////////////////////////////
// Step_Tick_Clock

type StepTickContext struct {

	// CurrentStep     *Step // 这里不提供具体的step, 以便适配不同的bank

	context *SequencerContext

	wktline *Timeline // the work-timeline

	// StepLength   Seconds // 每一步的时长
	// StepCount    uint    // 每一个 LOOP 包含的步数

	BreakCount uint // 中断计数器
	Broken     bool

	Info *SequencerInfo

	//  front|this|back|rtc

	RTC *RTC

	Front StepTickPos // 预先加载的位置
	Back  StepTickPos // 已经加载的位置
	This  StepTickPos // 正在加载的位置

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

	inst.innerUpdateCurrentPosition()
	if !inst.innerIsWalking() {
		return nil
	}

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

func (inst *innerSequencerTrigger) innerIsWalking() bool {
	ec := inst.context.EngineContext
	return ec.TransportContext.Walking()
}

func (inst *innerSequencerTrigger) innerUpdateCurrentPosition() error {

	ec := inst.context.EngineContext
	tc := ec.TransportContext
	current := ec.RTC.Precision

	if tc.Walking() {
		tc.current = current
		return nil
	}

	pos := tc.stopAt

	if pos != nil {
		sc := ec.SequencerContext
		current.Position = *pos
		current.Location = sc.ComputeStepLocation(pos)
		current.Index = sc.ComputeStepIndex(pos)
		tc.current = current
		return nil
	}

	return nil
}

// func (inst *innerSequencerTrigger) innerNormalizeSTP(stp *StepTickPos, stc *StepTickContext, align bool) error {

// 	const (
// 		min_count   = 4
// 		min_pos     = 0.001
// 		min_steplen = 0.001
// 	)

// 	pos := stp.Position.Pos
// 	steplen := stc.StepLength
// 	count := stc.StepCount

// 	if count < min_count {
// 		count = min_count
// 	}

// 	if pos < min_pos {
// 		pos = 0
// 	}

// 	if steplen < min_steplen {
// 		steplen = min_steplen
// 	}

// 	// for location
// 	location := StepLocation(pos / steplen)
// 	stp.Location = location

// 	// for index
// 	idx := (uint(location) % count)
// 	stp.Index = StepIndex(idx)

// 	// align
// 	if align {
// 		pos = (steplen * Seconds(location))
// 		stp.Position.Pos = pos
// 	}

// 	return nil
// }

func (inst *innerSequencerTrigger) innerDoPrepare(stc *StepTickContext) (*StepTickContext, error) {

	if stc == nil {
		stc = inst.context.STC
	}

	if stc == nil {
		stc = new(StepTickContext)
	}

	sc := inst.context
	ec := sc.EngineContext

	stc.wktline = ec.WorkTileline
	stc.context = sc

	return stc, nil
}

func (inst *innerSequencerTrigger) innerDoCheckParams(stc *StepTickContext) error {

	const (
		minStepLength    Seconds = 0.001
		minPreloadLength Seconds = 1
		minStepsPerCycle uint    = 4
	)

	if stc == nil {
		return fmt.Errorf("stc is nil")
	}

	ctx := stc.context
	if ctx == nil {
		return fmt.Errorf("stc.context is nil")
	}
	if ctx.EngineContext == nil {
		return fmt.Errorf("stc.context.EngineContext is nil")
	}
	if ctx.Sequencer == nil {
		return fmt.Errorf("stc.context.Sequencer is nil")
	}

	tline := stc.wktline
	if tline == nil {
		return fmt.Errorf("stc.wktline is nil")
	}

	info := stc.Info
	if info == nil {
		return fmt.Errorf("stc.Info is nil")
	}
	if stc.RTC == nil {
		return fmt.Errorf("stc.RTC is nil")
	}
	if info.StepsPerCycle < minStepsPerCycle {
		return fmt.Errorf("bad stc.Info.StepsPerCycle value: %v", info.StepsPerCycle)
	}
	if info.PreloadSeconds < minPreloadLength {
		return fmt.Errorf("bad stc.Info.PreloadSeconds value: %v", info.PreloadSeconds)
	}
	if info.SecondsPerStep < minStepLength {
		return fmt.Errorf("bad stc.Info.SecondsPerStep value: %v", info.SecondsPerStep)
	}

	stc.Front.Position.Ref = tline
	stc.This.Position.Ref = tline
	stc.Back.Position.Ref = tline

	return nil
}

// func (inst *innerSequencerTrigger) innerDoComputeRealtimePos(stc *StepTickContext, ce *ClockEvent) error {

// 	t1 := ce.Time
// 	wktlin := stc.wktline
// 	dst := &stc.Front
// 	abs := t1

// 	abs.Relative(wktlin, &dst.Position)

// 	const align = false
// 	return inst.innerNormalizeSTP(dst, stc, align)
// }

// func (inst *innerSequencerTrigger) innerDoComputeCurrentPos(stc *StepTickContext) error {

// 	src := &stc.Front // src
// 	dst := &stc.This  // dst

// 	dst.Position = src.Position

// 	const align = true
// 	return inst.innerNormalizeSTP(dst, stc, align)
// }

func (inst *innerSequencerTrigger) innerDoComputeAheadPos(stc *StepTickContext) error {

	info := stc.Info
	sc := stc.context
	ctx := sc.EngineContext
	pre_sec := info.PreloadSeconds

	sps := info.SecondsPerStep
	if sps < 0.001 {
		sps = 0.001
	}

	pre_step := int(pre_sec / sps)
	if pre_step < 1 {
		pre_step = 1
	}

	var front StepTickPos
	front = ctx.RTC.Normal

	front.Location += StepLocation(pre_step)
	front.Position.Pos += (sps * Seconds(pre_step))
	front.ComputeIndex(info, true)

	stc.Front = front
	return nil
}

func (inst *innerSequencerTrigger) innerDoFastForward(stc *StepTickContext) error {

	var current, from, to StepTickPos

	from = stc.Back
	to = stc.Front
	info := stc.Info
	max_gap := int(info.StepsPerCycle / 2)
	sps := info.SecondsPerStep

	current = from

	for current.Location < to.Location {

		gap := int(to.Location - current.Location)
		if gap > max_gap {
			stc.BreakCount++
			stc.Broken = true
			break
		}

		// dispatch clock
		stc.This = current
		err := inst.innerDispatchOnTickEvent(stc)
		handleError(err)

		// goto next
		current.Position.Pos += sps
		current.Location++
		current.ComputeIndex(info, true)
	}

	stc.Back = to
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
