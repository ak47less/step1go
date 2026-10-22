package core

import (
	"fmt"

	"github.com/ak47less/step1go/step1core"
)

type innerDefaultEngineFacade struct {
	context *step1core.EngineContext
}

// Join implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Join() error {
	panic("unimplemented")
}

// Route implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Route(bus *step1core.Bus) error {
	panic("unimplemented")
}

// Load implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Load() error {

	ctx := inst.context
	if ctx == nil {
		return fmt.Errorf("engine context is nil")
	}

	ls := ctx.LS
	if ls == nil {
		return fmt.Errorf("engine.LS is nil")
	}

	return ls.Load(inst.context)
}

// Reset implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Reset() error {
	panic("unimplemented")
}

// Configure implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Configure(ec *step1core.EngineContext, cfg *step1core.EngineConfiguration) error {
	// panic("unimplemented")

	return nil
}

// IO implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) IO(lc *step1core.Bus) error {
	// panic("unimplemented")

	return nil
}

// Clock implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Clock(lc *step1core.Bus) error {
	// panic("unimplemented")

	return nil
}

// Create implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Create() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Create()")

	return nil
}

// Destroy implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Destroy() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Destroy()")

	return nil
}

// GetConfiguration implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetConfiguration() *step1core.EngineConfiguration {
	return inst.context.Configuration
}

// GetContext implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetContext() *step1core.EngineContext {
	return inst.context
}

// GetDeviceManager implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetDeviceManager() step1core.DeviceManager {
	return inst.context.DeviceManager
}

// GetEngine implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetEngine() step1core.Engine {
	return inst.context.Engine
}

// GetInfo implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetInfo(ei *step1core.DeviceInfo) *step1core.DeviceInfo {
	panic("unimplemented")
}

// GetPorts implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetPorts() []step1core.Port {
	panic("unimplemented")
}

// GetProperties implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetProperties() []*step1core.Property {
	panic("unimplemented")
}

// GetTransport implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) GetTransport() step1core.Transport {
	return inst.context.Transport
}

// Init implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Init(cfg *step1core.EngineConfiguration) error {

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
func (inst *innerDefaultEngineFacade) Run() error {
	// panic("unimplemented")

	step1core.Log().Info("step1core.Engine.Run()")

	ctx := inst.context
	state := &ctx.State

	return state.WaitForStopping(-1)
}

// Start implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Start() error {

	step1core.Log().Info("step1core.Engine.Start()")

	lo := new(step1core.Looper)
	lo.Init(inst.context)

	inst.context.State.Starting = true
	inst.context.Looper = lo

	return lo.Start()
}

// Stop implements [step1core.Engine].
func (inst *innerDefaultEngineFacade) Stop() error {

	step1core.Log().Info("step1core.Engine.Stop()")

	inst.context.State.Stopping = true
	inst.context.Looper = nil

	return nil
}

func (inst *innerDefaultEngineFacade) _impl() step1core.Engine {
	return inst
}
