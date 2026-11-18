package lib

import (
	"fmt"

	"github.com/ak47less/step1go/lite2/core"
)

type EngineContext struct {
	AC *ApplicationContext

	Engine Engine

	Driver       Driver
	AudioContext AudioContext

	FrameTimeline    *core.Timeline
	LoopTimeline     *core.Timeline
	EngineTimeline   *core.Timeline
	DocumentTimeline *core.Timeline
	RootTimeline     *core.Timeline
}

type Engine interface {
	GetEngineContext() *EngineContext
}

////////////////////////////////////////////////////////////////////////////////

type DefaultEngine struct {
	context *EngineContext
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

	ec := inst.context
	audio := ec.AudioContext
	audio.Resume()

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
		Type: core.TimelineTypeLoop,
		Name: "@loop",
		Loop: true,
	}
	tl_frame := &core.Timeline{
		Type: core.TimelineTypeFrame,
		Name: "@frame",
		Loop: true,
	}

	tl_eng.Parent = tl_root
	tl_doc.Parent = tl_root
	tl_loop.Parent = tl_doc
	tl_frame.Parent = tl_doc

	ec.DocumentTimeline = tl_doc
	ec.EngineTimeline = tl_eng
	ec.FrameTimeline = tl_frame
	ec.LoopTimeline = tl_loop
	ec.RootTimeline = tl_root

}

func (inst *DefaultEngine) _impl() (Engine, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
