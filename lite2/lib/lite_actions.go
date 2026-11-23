package lib

type ActionName string

type Action struct {
	Name ActionName
}

type ActionEvent struct {
	Context *ActionContext
	Action  *Action
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
