package lite

import (
	"fmt"

	"github.com/ak47less/step1go/common"
)

////////////////////////////////////////////////////////////////////////////////

type innerActionHandlerFunc func(ec *EngineContext, ae *ActionEvent) error

type innerActionHandlerInfo struct {
	action  Action
	label   string
	handler innerActionHandlerFunc
}

////////////////////////////////////////////////////////////////////////////////

type ActionController interface {
	Controller

	Execute(ae *ActionEvent) error
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

	inst.addHandler(common.ActionMin, "example", hs.example)

	inst.addHandler(common.ActionPlay, "Play", hs.play)
	inst.addHandler(common.ActionStop, "Stop", hs.stop)
	inst.addHandler(common.ActionResume, "Resume", hs.resume)
	inst.addHandler(common.ActionPause, "Pause", hs.pause)
	inst.addHandler(common.ActionRecord, "Record", hs.record)

	inst.addHandler(common.ActionNoteOn, "NoteOn", hs.noteOn)
	inst.addHandler(common.ActionNoteOff, "NoteOff", hs.noteOff)

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
func (inst *innerActionController) Execute(ae *ActionEvent) error {

	action := ae.Action
	info := inst.table[action]

	if info == nil {
		return fmt.Errorf("no handler with action [%v]", action)
	}

	ec := inst.context
	fn := info.handler
	return fn(ec, ae)
}

// Start implements [ActionController].
func (inst *innerActionController) Start() {
	// nop
}

////////////////////////////////////////////////////////////////////////////////

type innerActionHandlerSet struct {
}

func (inst *innerActionHandlerSet) example(ec *EngineContext, ah *ActionEvent) error {
	return nil
}

func (inst *innerActionHandlerSet) play(ec *EngineContext, ah *ActionEvent) error {
	return ec.Transport.Play()
}

func (inst *innerActionHandlerSet) stop(ec *EngineContext, ah *ActionEvent) error {
	return ec.Transport.Stop()
}

func (inst *innerActionHandlerSet) record(ec *EngineContext, ah *ActionEvent) error {
	return ec.Transport.Record()
}

func (inst *innerActionHandlerSet) pause(ec *EngineContext, ah *ActionEvent) error {
	return ec.Transport.Pause()
}

func (inst *innerActionHandlerSet) resume(ec *EngineContext, ah *ActionEvent) error {
	return ec.Transport.Resume()
}

func (inst *innerActionHandlerSet) noteOn(ec *EngineContext, ah *ActionEvent) error {
	return fmt.Errorf("todo: no impl")
}

func (inst *innerActionHandlerSet) noteOff(ec *EngineContext, ah *ActionEvent) error {
	return fmt.Errorf("todo: no impl")
}

////////////////////////////////////////////////////////////////////////////////
// EOF
