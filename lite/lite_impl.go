package lite

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
	tl0 := new(Timeline)  // root
	tl1 := new(Timeline)  // work
	pt0 := new(TimePoint) // root
	pt1 := new(TimePoint) // work
	trans := new(innerTransport)
	timer := new(Timer)

	engine.context = ctx

	trans.ctx = ctx
	trans.tc = tc

	tl0.BPM = cfg.BPM
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

	inst.context.Timer = nil

	return nil
}

// Start implements [Engine].
func (inst *innerEngine) Start() error {

	timer := inst.context.Timer
	timer.Start()

	return nil
}

// Stop implements [Engine].
func (inst *innerEngine) Stop() error {

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

func (inst *innerTransport) ontimer(t *Timer) error {

	ctx := inst.ctx
	ac := ctx.AudioContext
	t0 := ac.GetCurrentTime()

	tp0 := ctx.CurrentAtRoot
	tp1 := ctx.CurrentAtWork
	off := tp1.At.GetTotalOffset()

	tp0.Position = t0
	tp1.Position = t0 - off

	return nil
}

// Create implements [Engine].
func (inst *innerTransport) _impl() Transport {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
