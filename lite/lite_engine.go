package lite

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

func handleError(err error) {

	if err == nil {
		return
	}
	log.Println(err.Error())

}

////////////////////////////////////////////////////////////////////////////////

type Timeline struct {
	Parent *Timeline

	Label string

	Offset Seconds

	Length Seconds

	BPM BPM

	Loop bool
}

func (inst *Timeline) HasSameRoot(other *Timeline) bool {
	r1 := inst.GetRoot()
	r2 := other.GetRoot()
	return (r1 == r2)
}

func (inst *Timeline) GetRoot() *Timeline {
	p := inst
	for ttl := 32; ttl > 0; ttl-- {
		if p != nil {
			parent := p.Parent
			if parent != nil {
				p = parent
				continue
			}
		}
		break
	}
	return p
}

func (inst *Timeline) GetTotalOffset() Seconds {
	p := inst
	var total Seconds = 0
	for ttl := 32; ttl > 0; ttl-- {
		if p == nil {
			break
		} else {
			total += p.Offset
			p = p.Parent
		}
	}
	return total
}

////////////////////////////////////////////////////////////////////////////////

type TimePoint struct {
	At       *Timeline
	Position Seconds
}

////////////////////////////////////////////////////////////////////////////////

type Timer struct {
	Delay    Seconds
	Interval Seconds
	Handler  func(t *Timer) error
	workerID int
}

func (inst *Timer) Start() {

	inst.workerID++

	wid := inst.workerID
	h := inst.Handler

	go func() {
		inst.work(wid, h)
	}()
}

func (inst *Timer) Stop() {
	inst.workerID = -1
}

func (inst *Timer) work(worker int, h func(t *Timer) error) {

	if h == nil {
		return
	}

	var err error
	t := inst.Delay

	for {

		du := inst.sec2du(t)
		time.Sleep(du)

		if worker != inst.workerID {
			break
		}

		err = h(inst)
		if err != nil {
			handleError(err)
		}

		t = inst.Interval
	}

}

func (inst *Timer) sec2du(sec Seconds) time.Duration {
	const min = time.Millisecond
	f := sec * Seconds(time.Second)
	return max(time.Duration(f), min)
}

////////////////////////////////////////////////////////////////////////////////

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
	Point   *TimePoint
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
	CurrentAtWork *TimePoint
	CurrentAtRoot *TimePoint

	State EngineState
}

type EngineContext = Context

////////////////////////////////////////////////////////////////////////////////
// EOF
