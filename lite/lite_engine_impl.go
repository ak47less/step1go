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
	lm := NewLifeManager()
	parts := NewPartContext(ctx)
	acfactory := driver.GetAudioContextFactory()
	ac := acfactory.NewAudioContext()
	tc := new(TransportContext)
	tl0 := new(Timeline)         // root
	tl1 := new(Timeline)         // work
	pt0 := new(RelativePosition) // root
	pt1 := new(RelativePosition) // work
	trans := new(innerTransport)
	timer := new(Timer)
	clock := new(SimpleClock)
	seqfactory := new(innerSequencerFactory)
	insfactory := cfg.InstrumentFactory
	ih := new(DefaultInstrumentHolder)
	profile := cfg.Profile
	im := NewInstrumentManager()

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

	pt0.Ref = tl0
	pt1.Ref = tl1

	ctx.LifeManager = lm
	ctx.PartManager = parts.Manager
	ctx.PartRegistry = parts.Registry
	ctx.Configuration = cfg
	ctx.Engine = engine
	ctx.AudioContext = ac
	ctx.AudioContextFactory = acfactory
	ctx.Transport = trans
	ctx.TransportContext = tc
	ctx.RootTileline = tl0
	ctx.WorkTileline = tl1
	ctx.Timer = timer
	ctx.TimerHandler = trans.ontimer
	ctx.Clock = clock
	ctx.Logger = Log()
	ctx.MainInstrumentFactory = insfactory
	ctx.CurrentInstrumentHolder = ih

	parts.Registry.RegisterPart(im)

	// Sequencer, rtc & stc

	seq, err := seqfactory.CreateSequencer(ctx)
	if err != nil {
		return nil, err
	}
	ctx.SequencerContext = seq.GetContext()

	sc := seq.GetContext()
	stc := sc.STC
	rtc := &ctx.RTC

	stc.RTC = rtc

	// init parts
	err = inst.innerInitParts(ctx)
	if err != nil {
		return nil, err
	}

	// init life-cycle
	err = inst.innerInitLifeCycle(ctx)
	if err != nil {
		return nil, err
	}

	// init Instruments
	err = inst.innerInitInstruments(ctx)
	if err != nil {
		return nil, err
	}

	return ctx.Engine, nil
}

func (inst *innerEngineFactory) innerInitInstruments(ec *Context) error {

	holder := ec.CurrentInstrumentHolder
	cube := &ec.SequencerContext.Cube
	tracks := cube.Tracks

	in1, err := holder.GetInstrument(ec)
	if err != nil {
		return err
	}

	for _, tr := range tracks {
		in2 := tr.Meta.Instrument
		if in2 == nil {
			in2 = in1
			tr.Meta.Instrument = in1
		}
	}

	return nil
}

func (inst *innerEngineFactory) innerInitParts(ec *Context) error {

	cfg := ec.Configuration
	src := cfg.Parts
	pr := ec.PartRegistry
	var err error

	for _, p := range src {

		part, ok := p.(Part)
		if ok {
			err = pr.RegisterPart(part)
		}

		factory, ok := p.(PartFactory)
		if ok {
			err = pr.RegisterFactory(factory)
		}

		registration, ok := p.(*PartRegistration)
		if ok {
			err = pr.Register(registration)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerEngineFactory) innerInitLifeCycle(ec *Context) error {

	src := ec.PartManager
	dst := ec.LifeManager
	namelist := src.ListNames()

	for _, name := range namelist {
		part, err := src.FindPartByName(name)
		if err != nil {
			return err
		}
		life, ok := part.(Life)
		if ok {
			dst.Add(life)
		}
	}
	return nil
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

// Walking implements [Transport].
func (inst *innerTransport) Walking() bool {
	return inst.tc.Walking()
}

// Pause implements [Transport].
func (inst *innerTransport) Pause() error {

	tc := inst.tc
	ctx := inst.ctx

	if tc.Walking() {
		at := ctx.RTC.Precision.Position.Clone()
		tc.stopAt = at
		tc.paused = true
	}

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
	tc.playing = true

	inst.innerTryRebaseToStopAt()

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

	inst.innerTryRebaseToStopAt()

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
	tc.paused = false

	inst.innerTryRebaseToStopAt()

	return inst.innerUpdateTransportContext(tc)
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {

	tc := inst.tc
	ctx := inst.ctx

	tc.recording = false
	tc.playing = false
	tc.paused = false

	at := new(RelativePosition)
	at.Ref = ctx.WorkTileline
	tc.stopAt = at

	return inst.innerUpdateTransportContext(tc)
}

func (inst *innerTransport) innerTryRebaseToStopAt() {

	at := inst.tc.stopAt
	inst.tc.stopAt = nil

	if at == nil {
		return
	}

	inst.ctx.RebaseToPosition(at)

	var tmp StepTickPos
	sc := inst.ctx.SequencerContext

	tmp.Position = *at
	tmp.Index = sc.ComputeStepIndex(at)
	tmp.Location = sc.ComputeStepLocation(at)

	sc.STC.Back = tmp

}

func (inst *innerTransport) innerUpdateTransportContext(tc *TransportContext) error {
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
	ctx.RebaseToPosition(to)
	return nil
}

func (inst *innerTransport) innerUpdateCurrentPos() error {

	// 这个方法的功能,目前由 Clock(RTC) 实现

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
	evt.Time = ctx.RTC.Precision.Position.Absolute()

	return clock.Dispatch(evt)
}

// Create implements [Engine].
func (inst *innerTransport) _impl() Transport {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type AutoEngineController struct {
	ec *EngineContext
}

// GetInfo implements [Part].
func (inst *AutoEngineController) GetInfo() *PartInfo {
	pi := &PartInfo{}
	pi.Name = "AutoEngineController"
	return pi
}

// GetLifeCycle implements [Life].
func (inst *AutoEngineController) GetLifeCycle() *LifeCycle {
	l := &LifeCycle{
		OnInit: inst.innerOnInit,

		OnCreate:  inst.innerOnCreate,
		OnStart:   inst.innerOnStart,
		OnRun:     inst.innerOnRun,
		OnStop:    inst.innerOnStop,
		OnDestroy: inst.innerOnDestroy,
	}
	return l
}

func (inst *AutoEngineController) innerOnInit(ec *EngineContext) error {
	inst.ec = ec
	return nil
}

func (inst *AutoEngineController) innerOnCreate() error {
	eng := inst.ec.Engine
	return eng.Create()
}

func (inst *AutoEngineController) innerOnStart() error {
	eng := inst.ec.Engine
	return eng.Start()
}

func (inst *AutoEngineController) innerOnRun() error {
	eng := inst.ec.Engine
	return eng.Join()
}

func (inst *AutoEngineController) innerOnStop() error {
	eng := inst.ec.Engine
	return eng.Stop()
}

func (inst *AutoEngineController) innerOnDestroy() error {
	eng := inst.ec.Engine
	return eng.Destroy()
}

func (inst *AutoEngineController) _impl() (Part, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
