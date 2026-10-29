package lite

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////

type innerActionHandlerFunc func(ec *EngineContext, ah *ActionHolder) error

type innerActionHandlerInfo struct {
	action  Action
	label   string
	handler innerActionHandlerFunc
}

////////////////////////////////////////////////////////////////////////////////

type ActionController interface {
	Controller

	Execute(ah *ActionHolder) error
}

func NewActionController(ec *EngineContext) ActionController {
	ctrl := new(innerActionController)
	ctrl.init(ec)
	return ctrl
}

////////////////////////////////////////////////////////////////////////////////

type innerActionController struct {
	context *EngineContext
	table   map[Action]*innerActionHandlerInfo
}

func (inst *innerActionController) init(ctx *EngineContext) {

	tab := make(map[Action]*innerActionHandlerInfo)
	hs := new(innerActionHandlerSet)

	inst.table = tab
	inst.context = ctx

	inst.addHandler(ActionExample, "Example", hs.example)

	inst.addHandler(ActionPlay, "Play", hs.play)
	inst.addHandler(ActionStop, "Stop", hs.stop)
	inst.addHandler(ActionResume, "Resume", hs.resume)
	inst.addHandler(ActionPause, "Pause", hs.pause)
	inst.addHandler(ActionRecord, "Record", hs.record)

}

func (inst *innerActionController) addHandler(action Action, label string, h innerActionHandlerFunc) {
	info := &innerActionHandlerInfo{
		action:  action,
		label:   label,
		handler: h,
	}
	inst.table[action] = info
}

// Execute implements [ActionController].
func (inst *innerActionController) Execute(ah *ActionHolder) error {

	action := ah.Action
	info := inst.table[action]

	if info == nil {
		return fmt.Errorf("no handler with action [%v]", action)
	}

	ec := inst.context
	fn := info.handler
	return fn(ec, ah)
}

// Start implements [ActionController].
func (inst *innerActionController) Start() {
	// nop
}

////////////////////////////////////////////////////////////////////////////////

type innerActionHandlerSet struct {
}

func (inst *innerActionHandlerSet) example(ec *EngineContext, ah *ActionHolder) error {
	return nil
}

func (inst *innerActionHandlerSet) play(ec *EngineContext, ah *ActionHolder) error {
	return ec.Transport.Play()
}

func (inst *innerActionHandlerSet) stop(ec *EngineContext, ah *ActionHolder) error {
	return ec.Transport.Stop()
}

func (inst *innerActionHandlerSet) record(ec *EngineContext, ah *ActionHolder) error {
	return ec.Transport.Record()
}

func (inst *innerActionHandlerSet) pause(ec *EngineContext, ah *ActionHolder) error {
	return ec.Transport.Pause()
}

func (inst *innerActionHandlerSet) resume(ec *EngineContext, ah *ActionHolder) error {
	return ec.Transport.Resume()
}

////////////////////////////////////////////////////////////////////////////////
// EOF
