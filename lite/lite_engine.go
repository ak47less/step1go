package lite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
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
}

////////////////////////////////////////////////////////////////////////////////

type ClockEvent struct {
	Clock   *Clock
	Time    *RelativePosition
	Context *Context
}

type ClockListener interface {
	OnClock(ce *ClockEvent) error
}

/**************************/
/*  class Clock           */
/**************************/

type Clock struct {
	mu        sync.RWMutex
	listeners []ClockListener
}

func (inst *Clock) AddListener(l ClockListener) {
	if isNilClockListener(l) {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	for _, item := range inst.listeners {
		if sameClockListener(item, l) {
			return
		}
	}
	inst.listeners = append(inst.listeners, l)
}

func (inst *Clock) RemoveListener(l ClockListener) {
	if isNilClockListener(l) {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	listeners := inst.listeners[:0]
	for _, item := range inst.listeners {
		if !sameClockListener(item, l) {
			listeners = append(listeners, item)
		}
	}
	inst.listeners = listeners
}

func (inst *Clock) Dispatch(ce *ClockEvent) error {
	if ce == nil {
		return errors.New("lite: nil clock event")
	}

	inst.mu.RLock()
	listeners := append([]ClockListener(nil), inst.listeners...)
	inst.mu.RUnlock()

	ce.Clock = inst

	var firstErr error
	var joinedErr error
	for _, listener := range listeners {
		err := listener.OnClock(ce)
		if err == nil {
			continue
		}
		if firstErr == nil {
			firstErr = err
		} else {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	if joinedErr != nil {
		return errors.Join(firstErr, joinedErr)
	}
	return firstErr
}

func isNilClockListener(l ClockListener) bool {
	if l == nil {
		return true
	}

	value := reflect.ValueOf(l)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func sameClockListener(a ClockListener, b ClockListener) bool {
	if isNilClockListener(a) || isNilClockListener(b) {
		return false
	}

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb || !ta.Comparable() {
		return false
	}
	return a == b
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

	CurrentInstrument Instrument

	MainInstrumentFactory InstrumentFactory

	Timer        *Timer
	TimerHandler func(*Timer) error

	RootTileline  *Timeline
	WorkTileline  *Timeline
	CurrentAtWork *RelativePosition
	CurrentAtRoot *RelativePosition

	State EngineState
}

type EngineContext = Context

////////////////////////////////////////////////////////////////////////////////
// EOF
