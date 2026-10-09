package core

import (
	"fmt"

	"github.com/ak47less/step1go/step1core"
)

type innerDefaultEngine struct {
	context *step1core.EngineContext
}

// Load implements [step1core.Engine].
func (inst *innerDefaultEngine) Load() error {

	ctx := inst.context
	if ctx == nil {
		return fmt.Errorf("engine context is nil")
	}

	loader := ctx.Loader
	if loader == nil {
		return fmt.Errorf("engine loader is nil")
	}

	return loader.Load(inst)
}

// Reset implements [step1core.Engine].
func (inst *innerDefaultEngine) Reset() error {
	panic("unimplemented")
}

// Configure implements [step1core.Engine].
func (inst *innerDefaultEngine) Configure(ec *step1core.EngineContext, cfg *step1core.EngineConfiguration) error {
	// panic("unimplemented")

	return nil
}

// IO implements [step1core.Engine].
func (inst *innerDefaultEngine) IO(lc *step1core.Bus) error {
	// panic("unimplemented")

	return nil
}

// Clock implements [step1core.Engine].
func (inst *innerDefaultEngine) Clock(lc *step1core.Bus) error {
	// panic("unimplemented")

	return nil
}

// Create implements [step1core.Engine].
func (inst *innerDefaultEngine) Create() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Create()")

	return nil
}

// Destroy implements [step1core.Engine].
func (inst *innerDefaultEngine) Destroy() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Destroy()")

	return nil
}

// GetConfiguration implements [step1core.Engine].
func (inst *innerDefaultEngine) GetConfiguration() *step1core.EngineConfiguration {
	return inst.context.Configuration
}

// GetContext implements [step1core.Engine].
func (inst *innerDefaultEngine) GetContext() *step1core.EngineContext {
	return inst.context
}

// GetElementManager implements [step1core.Engine].
func (inst *innerDefaultEngine) GetElementManager() step1core.ElementManager {
	return inst.context.ElementManager
}

// GetEngine implements [step1core.Engine].
func (inst *innerDefaultEngine) GetEngine() step1core.Engine {
	return inst.context.Engine
}

// GetInfo implements [step1core.Engine].
func (inst *innerDefaultEngine) GetInfo(ei *step1core.ElementInfo) *step1core.ElementInfo {
	panic("unimplemented")
}

// GetPorts implements [step1core.Engine].
func (inst *innerDefaultEngine) GetPorts() []step1core.Port {
	panic("unimplemented")
}

// GetProperties implements [step1core.Engine].
func (inst *innerDefaultEngine) GetProperties() []*step1core.Property {
	panic("unimplemented")
}

// GetTransport implements [step1core.Engine].
func (inst *innerDefaultEngine) GetTransport() step1core.Transport {
	return inst.context.Transport
}

// Init implements [step1core.Engine].
func (inst *innerDefaultEngine) Init(cfg *step1core.EngineConfiguration) error {

	ctx := inst.context
	if ctx == nil {
		ctx = new(step1core.EngineContext)
	}

	ctx.Configuration = cfg
	ctx.Logger = step1core.Log()

	inst.context = ctx
	return nil
}

// Run implements [step1core.Engine].
func (inst *innerDefaultEngine) Run() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Run()")

	ctx := inst.context
	state := &ctx.State

	return state.WaitForStopping(-1)
}

// Start implements [step1core.Engine].
func (inst *innerDefaultEngine) Start() error {

	step1core.Log().Info("step1core.Engine.Start()")

	lo := new(step1core.Looper)
	lo.Init(inst.context)

	inst.context.State.Starting = true
	inst.context.Looper = lo

	return lo.Start()
}

// Stop implements [step1core.Engine].
func (inst *innerDefaultEngine) Stop() error {

	step1core.Log().Info("step1core.Engine.Stop()")

	inst.context.State.Stopping = true
	inst.context.Looper = nil

	return nil
}

func (inst *innerDefaultEngine) _impl() step1core.Engine {
	return inst
}
