package core

import "github.com/ak47less/step1go"

type innerDefaultEngine struct {
	context *step1go.EngineContext
}

// Load implements [step1go.Engine].
func (inst *innerDefaultEngine) Load() error {
	panic("unimplemented")
}

// Reset implements [step1go.Engine].
func (inst *innerDefaultEngine) Reset() error {
	panic("unimplemented")
}

// Configure implements [step1go.Engine].
func (inst *innerDefaultEngine) Configure(ec *step1go.EngineContext, cfg *step1go.EngineConfiguration) error {
	// panic("unimplemented")

	return nil
}

// IO implements [step1go.Engine].
func (inst *innerDefaultEngine) IO(lc *step1go.Bus) error {
	// panic("unimplemented")

	return nil
}

// Clock implements [step1go.Engine].
func (inst *innerDefaultEngine) Clock(lc *step1go.Bus) error {
	// panic("unimplemented")

	return nil
}

// Create implements [step1go.Engine].
func (inst *innerDefaultEngine) Create() error {
	// panic("unimplemented")

	step1go.Log().Info("step1go.Engine.Create()")

	return nil
}

// Destroy implements [step1go.Engine].
func (inst *innerDefaultEngine) Destroy() error {
	// panic("unimplemented")

	step1go.Log().Info("step1go.Engine.Destroy()")

	return nil
}

// GetConfiguration implements [step1go.Engine].
func (inst *innerDefaultEngine) GetConfiguration() *step1go.EngineConfiguration {
	return inst.context.Configuration
}

// GetContext implements [step1go.Engine].
func (inst *innerDefaultEngine) GetContext() *step1go.EngineContext {
	return inst.context
}

// GetElementManager implements [step1go.Engine].
func (inst *innerDefaultEngine) GetElementManager() step1go.ElementManager {
	return inst.context.ElementManager
}

// GetEngine implements [step1go.Engine].
func (inst *innerDefaultEngine) GetEngine() step1go.Engine {
	return inst.context.Engine
}

// GetInfo implements [step1go.Engine].
func (inst *innerDefaultEngine) GetInfo(ei *step1go.ElementInfo) *step1go.ElementInfo {
	panic("unimplemented")
}

// GetPorts implements [step1go.Engine].
func (inst *innerDefaultEngine) GetPorts() []step1go.Port {
	panic("unimplemented")
}

// GetProperties implements [step1go.Engine].
func (inst *innerDefaultEngine) GetProperties() []*step1go.Property {
	panic("unimplemented")
}

// GetTransport implements [step1go.Engine].
func (inst *innerDefaultEngine) GetTransport() step1go.Transport {
	return inst.context.Transport
}

// Init implements [step1go.Engine].
func (inst *innerDefaultEngine) Init(cfg *step1go.EngineConfiguration) error {

	ctx := inst.context
	if ctx == nil {
		ctx = new(step1go.EngineContext)
	}

	ctx.Configuration = cfg
	ctx.Logger = step1go.Log()

	inst.context = ctx
	return nil
}

// Run implements [step1go.Engine].
func (inst *innerDefaultEngine) Run() error {
	// panic("unimplemented")

	step1go.Log().Info("step1go.Engine.Run()")

	ctx := inst.context
	state := &ctx.State

	return state.WaitForStopping(-1)
}

// Start implements [step1go.Engine].
func (inst *innerDefaultEngine) Start() error {

	step1go.Log().Info("step1go.Engine.Start()")

	lo := new(step1go.Looper)
	lo.Init(inst.context)

	inst.context.State.Starting = true
	inst.context.Looper = lo

	return lo.Start()
}

// Stop implements [step1go.Engine].
func (inst *innerDefaultEngine) Stop() error {

	step1go.Log().Info("step1go.Engine.Stop()")

	inst.context.State.Stopping = true
	inst.context.Looper = nil

	return nil
}

func (inst *innerDefaultEngine) _impl() step1go.Engine {
	return inst
}
