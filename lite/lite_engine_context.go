package lite

////////////////////////////////////////////////////////////////////////////////

// Context : the engine-Context
type Context struct {
	Configuration *Configuration

	// life & com

	Logger Logger

	Engine Engine

	LifeManager LifeManager

	// todo : comman comp-manager

	Transport Transport

	TransportContext *TransportContext

	SequencerContext *SequencerContext

	Clock *Clock

	AudioContext AudioContext

	AudioContextFactory AudioContextFactory

	CurrentInstrument       Instrument
	CurrentInstrumentHolder InstrumentHolder
	MainInstrumentFactory   InstrumentFactory

	Timer        *Timer
	TimerHandler func(*Timer) error

	RootTileline  *Timeline
	WorkTileline  *Timeline
	CurrentAtWork *RelativePosition
	CurrentAtRoot *RelativePosition

	State EngineState
}

func (inst *Context) GetPosAtWork(src *RelativePosition) *RelativePosition {

	if src == nil {
		src = new(RelativePosition)
	}

	if inst == nil {
		return src
	}

	line1 := src.At
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

type EngineContext = Context
