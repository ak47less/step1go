package lite

import (
	"log"
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
	Create() error
	Start() error
	Stop() error
	Destroy() error
}

type EngineFactory interface {
	CreateEngine(cfg *Configuration) (Engine, error)
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

type Clock struct {
}

func (inst *Clock) AddListener(l ClockListener) {

}

func (inst *Clock) RemoveListener(l ClockListener) {

}

func (inst *Clock) Dispatch(ce *ClockEvent) error {

}

////////////////////////////////////////////////////////////////////////////////

type Configuration struct {
	Driver Driver

	SampleRate int

	BPM BPM
}

type Context struct {
	Configuration *Configuration

	Engine Engine

	Transport Transport

	TransportContext *TransportContext

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
}

////////////////////////////////////////////////////////////////////////////////
// EOF
