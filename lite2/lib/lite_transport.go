package lib

import "github.com/ak47less/step1go/lite2/core"

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

func (inst *innerTransport) innerRebaseTo(rp *RelativePosition) error {

	if inst == nil || rp == nil {
		return nil
	}

	tc := inst.context
	ec := tc.AC.EngineContext

	want := rp.Absolute()
	have := inst.innerGetRawEnginePosition(ec.AC)

	tline := ec.Timelines.Transport
	diff := want - have
	old := tline.Offset

	tline.Offset = old + core.Seconds(diff)

	return nil
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
	ec := ctx.AC.EngineContext
	at1 := ec.Positions.StopAt
	tline := ec.Timelines.Document

	if at1 != nil {
		return nil
	}

	raw_pos := inst.innerGetRawEnginePosition(ctx.AC)

	ec.Positions.StopAt = raw_pos.Relative(tline, nil)
	ctx.Paused = true

	return nil
}

// Play implements [Transport].
func (inst *innerTransport) Play() error {

	ctx := inst.context
	ec := ctx.AC.EngineContext

	ctx.Playing = true
	ctx.Paused = false
	ctx.Recording = false

	at := ec.Positions.StopAt
	ec.Positions.StopAt = nil
	inst.innerRebaseTo(at)

	return nil
}

// Record implements [Transport].
func (inst *innerTransport) Record() error {

	ctx := inst.context
	ec := ctx.AC.EngineContext

	ctx.Playing = true
	ctx.Paused = false
	ctx.Recording = true

	at := ec.Positions.StopAt
	ec.Positions.StopAt = nil
	inst.innerRebaseTo(at)

	return nil
}

// Resume implements [Transport].
func (inst *innerTransport) Resume() error {

	tc := inst.context
	ec := tc.AC.EngineContext

	at := ec.Positions.StopAt
	ec.Positions.StopAt = nil

	inst.innerRebaseTo(at)
	tc.Paused = false

	return nil
}

// Seek implements [Transport].
func (inst *innerTransport) Seek(to *RelativePosition) error {
	return inst.innerRebaseTo(to)
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {

	ctx := inst.context
	ec := ctx.AC.EngineContext

	ctx.Playing = false
	ctx.Paused = false
	ctx.Recording = false

	// set stop_at

	tline := ec.Timelines.Document
	pos1 := ec.Positions.SelectionBegin
	pos2 := pos1.Relative(tline, nil)
	ec.Positions.StopAt = pos2

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
