package lite

import (
	"fmt"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// innerEngineFactory

func DefaultEngineFactory() EngineFactory {
	return new(innerEngineFactory)
}

type innerEngineFactory struct {
}

// CreateEngine implements [EngineFactory].
func (inst *innerEngineFactory) CreateEngine(cfg *Configuration) (Engine, error) {

	ctx := new(Context)
	engine := new(innerEngine)
	driver := cfg.Driver
	acfactory := driver.GetAudioContextFactory()
	ac := acfactory.NewAudioContext()
	tc := new(TransportContext)
	tl0 := new(Timeline)         // root
	tl1 := new(Timeline)         // work
	pt0 := new(RelativePosition) // root
	pt1 := new(RelativePosition) // work
	trans := new(innerTransport)
	timer := new(Timer)
	clock := new(Clock)
	seqfactory := new(innerSequencerFactory)
	insfactory := cfg.InstrumentFactory
	ii := new(InstrumentInfo)
	ih := new(DefaultInstrumentHolder)

	if insfactory == nil {
		return nil, fmt.Errorf("config.InstrumentFactory is nil")
	}

	engine.context = ctx

	trans.ctx = ctx
	trans.tc = tc

	tl0.Label = "root"
	tl0.BPM = cfg.BPM

	tl1.Label = "work"
	tl1.BPM = cfg.BPM
	tl1.Parent = tl0

	pt0.At = tl0
	pt1.At = tl1

	ctx.Configuration = cfg
	ctx.Engine = engine
	ctx.AudioContext = ac
	ctx.AudioContextFactory = acfactory
	ctx.Transport = trans
	ctx.TransportContext = tc
	ctx.RootTileline = tl0
	ctx.WorkTileline = tl1
	ctx.CurrentAtRoot = pt0
	ctx.CurrentAtWork = pt1
	ctx.Timer = timer
	ctx.TimerHandler = trans.ontimer
	ctx.Clock = clock
	ctx.Logger = Log()
	ctx.MainInstrumentFactory = insfactory
	ctx.CurrentInstrumentHolder = ih

	seq, err := seqfactory.CreateSequencer(ctx)
	if err != nil {
		return nil, err
	}
	ctx.SequencerContext = seq.GetContext()

	ii.Label = "default"
	ii.Name = "default"
	ii.Preset = "default"
	instrument, err := insfactory.CreateInstrument(ctx, ii)
	if err != nil {
		ac.Close()
		return nil, fmt.Errorf("create default instrument: %w", err)
	}
	ctx.CurrentInstrument = instrument

	return ctx.Engine, nil
}

func (inst *innerEngineFactory) _impl() EngineFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// innerEngine

type innerEngine struct {
	context *Context
}

// GetContext implements [Engine].
func (inst *innerEngine) GetContext() *EngineContext {
	return inst.context
}

// Join implements [Engine].
func (inst *innerEngine) Join() error {

	state := &inst.context.State

	if !state.Starting {
		return fmt.Errorf("call engine.start() first")
	}

	for {
		if state.Stopping || state.Stopped {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	return nil
}

// Create implements [Engine].
func (inst *innerEngine) Create() error {

	ctx := inst.context
	timer := ctx.Timer

	if timer == nil {
		timer = new(Timer)
		ctx.Timer = timer
	}

	timer.Delay = 0.01
	timer.Interval = 0.01
	timer.Handler = ctx.TimerHandler

	return nil
}

// Destroy implements [Engine].
func (inst *innerEngine) Destroy() error {

	ctx := inst.context

	ctx.Timer = nil
	ctx.State.Stopped = true

	return nil
}

// Start implements [Engine].
func (inst *innerEngine) Start() error {

	ctx := inst.context
	timer := ctx.Timer
	ac := ctx.AudioContext
	cfg := ctx.Configuration

	ctx.State.Starting = true
	ac.Resume()
	timer.Start()

	if cfg.Monitor != nil {
		mon := new(Monitor)
		mon.Init(ctx)
		mon.Start()
	}

	return nil
}

// Stop implements [Engine].
func (inst *innerEngine) Stop() error {

	inst.context.State.Stopping = true

	timer := inst.context.Timer
	timer.Stop()
	return nil
}

func (inst *innerEngine) _impl() Engine {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// innerEngine

type innerTransport struct {
	ctx *Context
	tc  *TransportContext
}

// Pause implements [Transport].
func (inst *innerTransport) Pause() error {
	tc := inst.tc
	tc.paused = true
	return nil
}

// Paused implements [Transport].
func (inst *innerTransport) Paused() bool {
	tc := inst.tc
	return tc.paused
}

// Play implements [Transport].
func (inst *innerTransport) Play() error {
	tc := inst.tc
	tc.playing = true
	return nil
}

// Playing implements [Transport].
func (inst *innerTransport) Playing() bool {
	tc := inst.tc
	return tc.playing
}

// Record implements [Transport].
func (inst *innerTransport) Record() error {
	tc := inst.tc
	tc.recording = true
	tc.playing = true
	return nil
}

// Recording implements [Transport].
func (inst *innerTransport) Recording() bool {
	tc := inst.tc
	return tc.recording
}

// Resume implements [Transport].
func (inst *innerTransport) Resume() error {
	tc := inst.tc
	tc.paused = false
	return nil
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {
	tc := inst.tc
	tc.recording = false
	tc.playing = false
	tc.paused = false
	return nil
}

func (inst *innerTransport) innerCheckUpdate() error {

	ctx := inst.ctx
	sc := ctx.SequencerContext

	if sc.Updating {
		seq := sc.Sequencer
		return seq.Update(sc)
	}

	return nil
}

func (inst *innerTransport) ontimer(t *Timer) error {

	err := inst.innerCheckUpdate()
	if err != nil {
		return err
	}

	ctx := inst.ctx
	ac := ctx.AudioContext
	t0 := ac.GetCurrentTime()
	clock := ctx.Clock
	evt := new(ClockEvent)

	ctx.State.Started = true

	tp0 := ctx.CurrentAtRoot
	tp1 := ctx.CurrentAtWork
	off := tp1.At.GetTotalOffset()

	tp0.Position = t0
	tp1.Position = t0 - off

	evt.Clock = clock
	evt.Context = ctx
	evt.Time = tp1

	return clock.Dispatch(evt)
}

// Create implements [Engine].
func (inst *innerTransport) _impl() Transport {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
