package lite

import (
	"fmt"
	"log"
)

////////////////////////////////////////////////////////////////////////////////

func handleError(err error) {

	if err == nil {
		return
	}
	log.Println(err.Error())

}

////////////////////////////////////////////////////////////////////////////////
// Engine

type Engine interface {
	GetContext() *EngineContext

	Create() error

	Start() error

	Join() error

	Stop() error

	Destroy() error
}

type EngineFactory interface {
	CreateEngine(cfg *Configuration) (Engine, error)
}

type EngineState struct {
	Starting bool
	Stopping bool
	Stopped  bool
	Started  bool
}

func (inst *EngineState) String() string {
	if inst == nil {
		return ""
	}
	s1 := inst.Starting
	s2 := inst.Started
	s3 := inst.Stopping
	s4 := inst.Stopped
	return fmt.Sprintf("engine.[state starting:%v started:%v stopping:%v stopped:%v]", s1, s2, s3, s4)
}

////////////////////////////////////////////////////////////////////////////////

type Transport interface {
	Play() error
	Record() error
	Stop() error
	Pause() error
	Resume() error

	Paused() bool
	Recording() bool
	Playing() bool
}

type TransportContext struct {
	paused    bool
	playing   bool
	recording bool

	running bool // 综合各个状态, 确定是否正在走带

	rebaseToNewPosition *RelativePosition // 如果这个字段不为nil, 表示需要重定向到新的位置 @work.timeline
}

////////////////////////////////////////////////////////////////////////////////

// Context : the engine-Context
type Context struct {
	Configuration *Configuration

	Logger Logger

	Engine Engine

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

type EngineContext = Context

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

////////////////////////////////////////////////////////////////////////////////
// EOF
