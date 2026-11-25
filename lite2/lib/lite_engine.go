package lib

import (
	"fmt"
	"time"

	"github.com/ak47less/step1go/lite2/core"
)

type Timelines struct {
	Frame     *core.Timeline
	Loop      *core.Timeline
	Engine    *core.Timeline
	Document  *core.Timeline
	Step      *core.Timeline
	Root      *core.Timeline
	Transport *core.Timeline
}

type Positions struct {
	Current        AbsolutePosition
	Preload        AbsolutePosition
	SelectionBegin AbsolutePosition
	SelectionEnd   AbsolutePosition

	Raw    *RelativePosition
	StopAt *RelativePosition
}

type EngineContext struct {
	AC *ApplicationContext

	Engine Engine

	Driver       Driver
	AudioContext AudioContext

	// 用一个乐器作为MIDI输入接口
	Instrument       Instrument
	InstrumentHolder InstrumentHolder

	Positions Positions

	Timelines Timelines

	Starting bool
	Started  bool
	Stopping bool
	Stopped  bool
}

type Engine interface {
	GetEngineContext() *EngineContext

	GetRawPosition(dst *RelativePosition)

	Start() error

	Stop() error

	Restart() error
}

////////////////////////////////////////////////////////////////////////////////

type DefaultEngine struct {
	context *EngineContext
}

// GetRawPosition implements [Engine].
func (inst *DefaultEngine) GetRawPosition(dst *RelativePosition) {

	if inst == nil || dst == nil {
		return
	}

	ec := inst.context

	dst.Pos = ec.AudioContext.GetCurrentTime()
	dst.Ref = ec.Timelines.Engine
}

// Restart implements [Engine].
func (inst *DefaultEngine) Restart() error {

	inst.Stop()

	ctx := inst.context
	ctx.Starting = false
	ctx.Started = false
	ctx.Stopping = false
	ctx.Stopped = false

	return inst.Start()
}

func (inst *DefaultEngine) innerStart() error {

	ec := inst.context
	ec.Starting = true

	audio := ec.AudioContext
	audio.Resume()

	return nil
}

func (inst *DefaultEngine) innerStop() error {

	ec := inst.context
	ec.Stopping = true

	return nil
}

// Start implements [Engine].
func (inst *DefaultEngine) Start() error {
	return inst.innerStart()
}

// Stop implements [Engine].
func (inst *DefaultEngine) Stop() error {
	return inst.innerStop()
}

// GetEngineContext implements [Engine].
func (inst *DefaultEngine) GetEngineContext() *EngineContext {
	return inst.context
}

// GetLifeCycle implements [Life].
func (inst *DefaultEngine) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit:   inst.onInit,
		OnCreate: inst.onCreate,
		OnStart:  inst.onStart,
		OnLoop:   inst.innerOnLoop,
		OnStop:   inst.onStop,
	}
	return lc
}

func (inst *DefaultEngine) onInit(ac *ApplicationContext) error {

	ec := ac.EngineContext
	ec.Engine = inst
	ac.Engine = inst

	inst.context = ec
	inst.innerInitTimelines()

	return nil
}

func (inst *DefaultEngine) onCreate() error {

	return inst.innerLoadDriver()
}

func (inst *DefaultEngine) onStart() error {
	return inst.innerStart()
}

func (inst *DefaultEngine) onStop() error {
	return inst.innerStop()
}

func (inst *DefaultEngine) innerOnLoop() error {

	const interval = time.Second / 10
	ec := inst.context.AC.EngineContext

	ec.Started = true
	defer func() {
		ec.Stopped = true
	}()

	for {
		if ec.Stopping {
			break
		}
		if ec.Stopped {
			break
		}

		time.Sleep(interval)
	}
	return nil
}

func (inst *DefaultEngine) innerLoadDriver() error {

	ec := inst.context
	cfg := ec.AC.Configuration
	driver := cfg.Driver

	if driver == nil {
		return fmt.Errorf("no audio driver been configured")
	}

	factory := driver.GetFactory()
	audio, err := factory.CreateAudioContext()

	if err != nil {
		return err
	}

	ec.Driver = driver
	ec.AudioContext = audio

	return nil
}

func (inst *DefaultEngine) innerInitTimelines() {

	ec := inst.context
	tls := &ec.Timelines

	tl_eng := &core.Timeline{
		Type: core.TimelineTypeEngine,
		Name: "@engine",
	}
	tl_doc := &core.Timeline{
		Type: core.TimelineTypeDocument,
		Name: "@document",
	}
	tl_root := &core.Timeline{
		Type: core.TimelineTypeRoot,
		Name: "@root",
	}
	tl_loop := &core.Timeline{
		Type:   core.TimelineTypeLoop,
		Name:   "@loop",
		Loop:   true,
		Length: 1,
	}
	tl_frame := &core.Timeline{
		Type:   core.TimelineTypeFrame,
		Name:   "@frame",
		Loop:   true,
		Length: 1,
	}
	tl_trans := &core.Timeline{
		Type: core.TimelineTypeTransport,
		Name: "@transport",
	}
	tl_step := &core.Timeline{
		Type:   core.TimelineTypeStep,
		Name:   "@step",
		Loop:   true,
		Length: 1,
	}

	tl_eng.Parent = tl_trans
	tl_doc.Parent = tl_root
	tl_trans.Parent = tl_root
	tl_loop.Parent = tl_doc
	tl_frame.Parent = tl_doc
	tl_step.Parent = tl_doc
	tl_root.Parent = nil

	tls.Document = tl_doc
	tls.Engine = tl_eng
	tls.Frame = tl_frame
	tls.Loop = tl_loop
	tls.Root = tl_root
	tls.Step = tl_step
	tls.Transport = tl_trans

}

func (inst *DefaultEngine) _impl() (Engine, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
