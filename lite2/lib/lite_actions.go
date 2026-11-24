package lib

type ActionName string

type Action struct {
	Context *ActionContext
	Name    ActionName
}

type ActionEvent struct {
	Context *ActionContext
	Action  *Action
	Sender  any
}

type ActionContext struct {
	AC *ApplicationContext
	EC *EngineContext
}

type ActionController interface {
	NewContext() *ActionContext

	StartEngine(ctx *ActionContext) error
	StopEngine(ctx *ActionContext) error
	RestartEngine(ctx *ActionContext) error

	Play(ctx *ActionContext) error
	Stop(ctx *ActionContext) error
	Pause(ctx *ActionContext) error
	Resume(ctx *ActionContext) error
	Record(ctx *ActionContext) error

	Seek(ctx *ActionContext, pos *RelativePosition) error
}

type ActionEventHandler interface {
	Handle(ae *ActionEvent) error
}

////////////////////////////////////////////////////////////////////////////////

type DefaultActionController struct {
	context *ApplicationContext
}

// NewContext implements [ActionController].
func (inst *DefaultActionController) NewContext() *ActionContext {
	ctx := inst.context
	dst := &ActionContext{
		AC: ctx,
		EC: ctx.EngineContext,
	}
	return dst
}

// Pause implements [ActionController].
func (inst *DefaultActionController) Pause(ctx *ActionContext) error {
	tr := inst.context.Transport
	return tr.Pause()
}

// Play implements [ActionController].
func (inst *DefaultActionController) Play(ctx *ActionContext) error {
	tr := inst.context.Transport
	return tr.Play()
}

// Record implements [ActionController].
func (inst *DefaultActionController) Record(ctx *ActionContext) error {
	tr := inst.context.Transport
	return tr.Record()
}

// RestartEngine implements [ActionController].
func (inst *DefaultActionController) RestartEngine(ctx *ActionContext) error {
	inst.StopEngine(ctx)
	return inst.StartEngine(ctx)
}

// Resume implements [ActionController].
func (inst *DefaultActionController) Resume(ctx *ActionContext) error {
	tr := inst.context.Transport
	return tr.Resume()
}

// Seek implements [ActionController].
func (inst *DefaultActionController) Seek(ctx *ActionContext, pos *RelativePosition) error {
	panic("unimplemented")
}

// StartEngine implements [ActionController].
func (inst *DefaultActionController) StartEngine(ctx *ActionContext) error {
	engine := ctx.EC.Engine
	return engine.Start()
}

// Stop implements [ActionController].
func (inst *DefaultActionController) Stop(ctx *ActionContext) error {
	tr := inst.context.Transport
	return tr.Stop()
}

// StopEngine implements [ActionController].
func (inst *DefaultActionController) StopEngine(ctx *ActionContext) error {
	engine := ctx.EC.Engine
	return engine.Stop()
}

// GetLifeCycle implements [Life].
func (inst *DefaultActionController) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit:   inst.innerOnInit,
		OnCreate: inst.innerOnCreate,
	}
	return lc
}

func (inst *DefaultActionController) innerOnInit(ac *ApplicationContext) error {
	inst.context = ac
	ac.ActionController = inst
	return nil
}

func (inst *DefaultActionController) innerOnCreate() error {
	return nil
}

func (inst *DefaultActionController) _impl() (Life, ActionController) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
