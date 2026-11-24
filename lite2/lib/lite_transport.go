package lib

type Transport interface {
	Play() error
	Record() error
	Stop() error
	Pause() error
	Resume() error

	Seek(to *RelativePosition) error

	GetCurrentPosition() *RelativePosition
}

type TransportContext struct {
	AC *ApplicationContext

	Transport Transport

	Playing   bool
	Recording bool
	Paused    bool
}

////////////////////////////////////////////////////////////////////////////////

func NewTransport() Transport {
	return new(innerTransport)
}

////////////////////////////////////////////////////////////////////////////////

type innerTransport struct {
	context *TransportContext
}

func (inst *innerTransport) innerGetRawEnginePosition(appc *ApplicationContext) AbsolutePosition {

	ec := appc.EngineContext
	raw_sec := ec.AudioContext.GetCurrentTime()
	tls := &ec.Timelines

	var rp1 RelativePosition
	rp1.Ref = tls.Engine
	rp1.Pos = raw_sec

	return rp1.Absolute()
}

// GetCurrentPosition implements [Transport].
func (inst *innerTransport) GetCurrentPosition() *RelativePosition {

	tc := inst.context
	ec := tc.AC.EngineContext
	at := ec.Positions.StopAt
	tls := &ec.Timelines
	tline := tls.Document
	var abs AbsolutePosition

	if at != nil {
		abs = at.Absolute()
	} else {
		abs = inst.innerGetRawEnginePosition(ec.AC)
	}

	return abs.Relative(tline, nil)
}

// Pause implements [Transport].
func (inst *innerTransport) Pause() error {

	ctx := inst.context
	ctx.Paused = true

	return nil
}

// Play implements [Transport].
func (inst *innerTransport) Play() error {

	ctx := inst.context
	ctx.Playing = true
	ctx.Paused = false
	ctx.Recording = false

	return nil
}

// Record implements [Transport].
func (inst *innerTransport) Record() error {

	ctx := inst.context
	ctx.Playing = true
	ctx.Paused = false
	ctx.Recording = true

	return nil
}

// Resume implements [Transport].
func (inst *innerTransport) Resume() error {

	ctx := inst.context
	ctx.Paused = false

	return nil
}

// Seek implements [Transport].
func (inst *innerTransport) Seek(to *RelativePosition) error {
	panic("unimplemented")
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {

	ctx := inst.context
	ctx.Playing = false
	ctx.Paused = false
	ctx.Recording = false

	return nil
}

// GetLifeCycle implements [Life].
func (inst *innerTransport) GetLifeCycle() *LifeCycle {
	return &LifeCycle{
		Order:  0,
		OnInit: inst.innerOnInit,
	}
}

func (inst *innerTransport) innerOnInit(ac *ApplicationContext) error {

	ac.TransportContext.Transport = inst
	ac.Transport = inst

	inst.context = ac.TransportContext

	return nil
}

func (inst *innerTransport) _impl() (Transport, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
