package lite

////////////////////////////////////////////////////////////////////////////////

// Context : the engine-Context
type EngineContext struct {

	// api

	Transport Transport

	LifeManager LifeManager

	PartManager PartManager

	PartRegistry PartRegistry

	Logger Logger

	Engine Engine

	Clock Clock

	EventDispatcher EventDispatcher

	// sub-context

	TransportContext *TransportContext

	SequencerContext *SequencerContext

	// config

	Configuration *Configuration

	// life & com

	AudioContext AudioContext

	AudioContextFactory AudioContextFactory

	CurrentInstrument       Instrument
	CurrentInstrumentHolder InstrumentHolder
	MainInstrumentFactory   InstrumentFactory
	InstrumentManager       InstrumentManager

	Timer        *Timer
	TimerHandler func(*Timer) error

	RootTileline *Timeline
	WorkTileline *Timeline

	RTC RTC

	State EngineState
}

func (inst *EngineContext) GetPosAtWork(src *RelativePosition) *RelativePosition {

	if src == nil {
		src = new(RelativePosition)
	}

	if inst == nil {
		return src
	}

	line1 := src.Ref
	line2 := inst.WorkTileline

	if line1 == nil || line2 == nil {
		return src
	}

	if line1.IsTypeOf(TimelineTypeWork) {
		return src
	}

	abs := src.Absolute()
	dst := new(RelativePosition)
	return abs.Relative(line2, dst)
}

func (inst *EngineContext) GetCurrentPosition(dst *RelativePosition) *RelativePosition {

	if dst == nil {
		dst = new(RelativePosition)
	}

	if inst == nil {
		return dst
	}

	at := inst.TransportContext.stopAt

	if at == nil {
		at = &inst.RTC.Precision.Position
	}

	*dst = *at
	return dst
}

func (inst *EngineContext) RebaseToPosition(to *RelativePosition) {

	if inst == nil || to == nil {
		return
	}

	ctx := inst
	work := ctx.WorkTileline
	abs := to.Absolute()
	to2 := abs.Relative(work, nil)

	current_work_pos := to2.Pos
	current_root_pos := ctx.RTC.Raw.Position.Absolute()

	// work.offset + current_work_pos = current_root_pos

	work.Offset = Seconds(current_root_pos) - current_work_pos
}

type Context = EngineContext
