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
	ih := new(DefaultInstrumentHolder)
	profile := cfg.Profile

	if profile == nil {
		profile = (new(SimpleProfile)).Init()
		cfg.Profile = profile
	}

	if insfactory == nil {
		return nil, fmt.Errorf("config.InstrumentFactory is nil")
	}

	engine.context = ctx

	trans.ctx = ctx
	trans.tc = tc

	tl0.Label = "root"
	tl0.Profile = profile

	tl1.Label = "work"
	tl1.Profile = profile
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
	return inst.innerUpdateTransportContext(tc)
}

// Paused implements [Transport].
func (inst *innerTransport) Paused() bool {
	tc := inst.tc
	return tc.paused
}

// Play implements [Transport].
func (inst *innerTransport) Play() error {

	tc := inst.tc
	current := inst.ctx.CurrentAtWork.Clone()

	tc.playing = true
	tc.rebaseToNewPosition = current

	return inst.innerUpdateTransportContext(tc)
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

	return inst.innerUpdateTransportContext(tc)
}

// Recording implements [Transport].
func (inst *innerTransport) Recording() bool {
	tc := inst.tc
	return tc.recording
}

// Resume implements [Transport].
func (inst *innerTransport) Resume() error {

	tc := inst.tc
	current := inst.ctx.CurrentAtWork.Clone()

	tc.paused = false
	tc.rebaseToNewPosition = current

	return inst.innerUpdateTransportContext(tc)
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {

	tc := inst.tc
	tc.recording = false
	tc.playing = false
	tc.paused = false

	return inst.innerUpdateTransportContext(tc)
}

func (inst *innerTransport) innerUpdateTransportContext(tc *TransportContext) error {
	if tc != nil {
		play := tc.playing
		pause := tc.paused
		tc.running = (!pause) && play
	}
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

// 重新确定基准位置
func (inst *innerTransport) innerRebaseCurrentPos(ctx *Context, to *RelativePosition) error {

	pos := to.Position
	line_wk := ctx.WorkTileline
	current_pos_abs := ctx.CurrentAtRoot.Absolute()
	current_pos_wk := ctx.CurrentAtWork

	current_pos_wk.At = line_wk
	current_pos_wk.Position = pos
	line_wk.Offset = Seconds(current_pos_abs) - pos

	return nil
}

func (inst *innerTransport) innerUpdateCurrentPos() error {

	ctx := inst.ctx
	ac := ctx.AudioContext
	t0 := ac.GetCurrentTime()
	tc := ctx.TransportContext
	pos_at_root := ctx.CurrentAtRoot
	pos_at_work := ctx.CurrentAtWork
	rebase := tc.rebaseToNewPosition

	pos_at_root.Position = t0

	if tc.running {
		off := pos_at_work.At.GetTotalOffset()
		pos_at_work.Position = t0 - off
	}

	if rebase != nil {
		tc.rebaseToNewPosition = nil
		inst.innerRebaseCurrentPos(ctx, rebase)
	}

	return nil
}

func (inst *innerTransport) ontimer(t *Timer) error {

	err := inst.innerCheckUpdate()
	if err != nil {
		return err
	}

	inst.innerUpdateCurrentPos()

	ctx := inst.ctx
	clock := ctx.Clock
	evt := new(ClockEvent)

	ctx.State.Started = true

	evt.Clock = clock
	evt.Context = ctx
	evt.Time = ctx.CurrentAtWork

	return clock.Dispatch(evt)
}

// Create implements [Engine].
func (inst *innerTransport) _impl() Transport {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
