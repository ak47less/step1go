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
	panic("unimplemented")
}

// Play implements [ActionController].
func (inst *DefaultActionController) Play(ctx *ActionContext) error {
	panic("unimplemented")
}

// Record implements [ActionController].
func (inst *DefaultActionController) Record(ctx *ActionContext) error {
	panic("unimplemented")
}

// RestartEngine implements [ActionController].
func (inst *DefaultActionController) RestartEngine(ctx *ActionContext) error {
	panic("unimplemented")
}

// Resume implements [ActionController].
func (inst *DefaultActionController) Resume(ctx *ActionContext) error {
	panic("unimplemented")
}

// Seek implements [ActionController].
func (inst *DefaultActionController) Seek(ctx *ActionContext, pos *RelativePosition) error {
	panic("unimplemented")
}

// StartEngine implements [ActionController].
func (inst *DefaultActionController) StartEngine(ctx *ActionContext) error {
	panic("unimplemented")
}

// Stop implements [ActionController].
func (inst *DefaultActionController) Stop(ctx *ActionContext) error {
	panic("unimplemented")
}

// StopEngine implements [ActionController].
func (inst *DefaultActionController) StopEngine(ctx *ActionContext) error {
	panic("unimplemented")
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
