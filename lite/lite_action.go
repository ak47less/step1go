package lite

import (
	"sync"

	"github.com/ak47less/step1go/common"
)

////////////////////////////////////////////////////////////////////////////////
// Action (id)

type Action int

const (
	ActionMin Action = iota

	ActionPlay
	ActionStop
	ActionPause
	ActionResume
	ActionRecord

	ActionStartEngine
	ActionStopEngine

	ActionNoteOn
	ActionNoteOff

	ActionKeyUp
	ActionKeyDown

	ActionExample
	ActionMax
)

var theActionStringifyTable innerActionStringifyTable

func (a Action) String() string {
	tab := &theActionStringifyTable
	return tab.stringify(a)
}

////////////////////////////////////////////////////////////////////////////////

type innerActionStringifyTable struct {
	table map[Action]string
	mu    sync.Mutex
}

func (inst *innerActionStringifyTable) init() {

	tab := inst.table
	if tab != nil {
		return
	}

	tab = make(map[Action]string)

	tab[ActionExample] = "example"
	tab[ActionMin] = "min"
	tab[ActionMax] = "max"

	tab[ActionKeyUp] = "key_up"
	tab[ActionKeyDown] = "key_down"

	tab[ActionNoteOff] = "note_off"
	tab[ActionNoteOn] = "note_on"

	tab[ActionStartEngine] = "engine_start"
	tab[ActionStopEngine] = "engine_stop"

	tab[ActionPlay] = "play"
	tab[ActionStop] = "stop"
	tab[ActionRecord] = "record"
	tab[ActionResume] = "resume"
	tab[ActionPause] = "pause"

	inst.table = tab
}

func (inst *innerActionStringifyTable) stringify(a Action) string {

	mu := &inst.mu
	mu.Lock()
	defer mu.Unlock()

	inst.init()
	str := inst.table[a]
	if str == "" {
		str = "[undef_action]"
	}
	return str
}

////////////////////////////////////////////////////////////////////////////////
// ActionHolder

type ActionHolder struct {
	Action Action

	KeyAction *KeyAction

	NoteAction *NoteAction
}

func (inst *ActionHolder) String() string {
	if inst == nil {
		return ""
	}
	return inst.Action.String()
}

////////////////////////////////////////////////////////////////////////////////
// KeyAction

type KeyCode rune

type KeyAction struct {
	Action  Action
	KeyName string
	KeyCode KeyCode
	KeyUp   bool
	KeyDown bool
}

func (inst *KeyAction) Clone() *KeyAction {
	if inst == nil {
		return nil
	}
	k2 := new(KeyAction)
	*k2 = *inst
	return k2
}

////////////////////////////////////////////////////////////////////////////////
// NoteAction

type NoteAction struct {
	Action   Action
	Note     common.MidiNoteNumber
	Velocity common.MidiNoteVelocity
}

////////////////////////////////////////////////////////////////////////////////
// EOF
