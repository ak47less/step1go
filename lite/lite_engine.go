package lite

import (
	"fmt"

	"github.com/ak47less/step1go/common"
)

////////////////////////////////////////////////////////////////////////////////

func handleError(err error) {

	if err == nil {
		return
	}

	// log.Println(err.Error())

	common.Log().Error("%s", err.Error())

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

	Walking() bool
}

////////////////////////////////////////////////////////////////////////////////

type TransportContext struct {
	paused    bool
	playing   bool
	recording bool

	current StepTickPos // current@work

	stopAt *RelativePosition // 如果这个字段不为nil, 表示当前位置停止在这个地方 @work.timeline
}

func (inst *TransportContext) Walking() bool {
	if inst == nil {
		return false
	}
	return (inst.playing) && (!inst.paused)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
