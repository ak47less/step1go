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

	AudioContext AudioContext

	AudioContextFactory AudioContextFactory

	CurrentInstrument Instrument

	MainInstrumentFactory InstrumentFactory

	RootTileline *Timeline

	WorkTileline *Timeline
}

////////////////////////////////////////////////////////////////////////////////
// EOF
