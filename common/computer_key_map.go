package common

import (
	"fmt"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type ComputerKeyURI string

////////////////////////////////////////////////////////////////////////////////
// ComputerKeyMap

type ComputerKeyMap struct {
	table map[ComputerKeyURI]*ComputerKeyMapping
	mu    sync.Mutex
}

func (inst *ComputerKeyMap) Add(item *ComputerKeyMapping) {
	const rewrite = true
	if inst == nil || item == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.init()
	u := item.ComputeURI(rewrite)
	inst.table[u] = item
}

func (inst *ComputerKeyMap) init() *ComputerKeyMap {
	if inst == nil {
		inst = new(ComputerKeyMap)
	}
	if inst.table == nil {
		inst.table = make(map[ComputerKeyURI]*ComputerKeyMapping)
	}
	return inst
}

func (inst *ComputerKeyMap) Find(want *ComputerKeyMapping) (have *ComputerKeyMapping, err error) {
	if inst == nil {
		return nil, fmt.Errorf("computer key map is nil")
	}
	if want == nil {
		return nil, fmt.Errorf("computer key mapping is nil")
	}

	const rewrite = false
	uri := want.ComputeURI(rewrite)

	inst.mu.Lock()
	defer inst.mu.Unlock()

	have = inst.table[uri]
	if have == nil {
		return nil, fmt.Errorf("computer key mapping not found: %q", uri)
	}
	return have, nil
}

////////////////////////////////////////////////////////////////////////////////

type ComputerKeyMapBuilder struct {
	items []*ComputerKeyMapping

	code ComputerKeyCode
	name ComputerKeyName
}

func (inst *ComputerKeyMapBuilder) Build() *ComputerKeyMap {
	result := new(ComputerKeyMap)
	if inst == nil {
		return result
	}
	for _, item := range inst.items {
		result.Add(item)
	}
	return result
}

func (inst *ComputerKeyMapBuilder) MapKeyName(name ComputerKeyName) *ComputerKeyMapBuilder {
	if inst == nil {
		return nil
	}
	inst.name = name
	inst.code = CKeyCodeMin
	return inst
}

func (inst *ComputerKeyMapBuilder) MapKeyCode(code ComputerKeyCode) *ComputerKeyMapBuilder {
	if inst == nil {
		return nil
	}
	inst.code = code
	if code > CKeyCodeMin && code < CKeyCodeMax {
		inst.name = code.Name()
	} else {
		inst.name = ""
	}
	return inst
}

func (inst *ComputerKeyMapBuilder) ToNote(note MidiNoteNumber) {
	if inst == nil || !inst.hasKey() {
		return
	}

	const velocity MidiNoteVelocity = 127
	inst.items = append(inst.items,
		inst.newMapping(false, true, &ActionEvent{
			Action:       ActionNoteOn,
			FlagNote:     true,
			FlagNoteOn:   true,
			NoteNumber:   note,
			NoteVelocity: velocity,
		}),
		inst.newMapping(true, false, &ActionEvent{
			Action:       ActionNoteOff,
			FlagNote:     true,
			FlagNoteOff:  true,
			NoteNumber:   note,
			NoteVelocity: velocity,
		}),
	)
}

func (inst *ComputerKeyMapBuilder) ToControl(action Action) {
	if inst == nil || !inst.hasKey() {
		return
	}
	inst.items = append(inst.items, inst.newMapping(false, true, &ActionEvent{
		Action:      action,
		FlagControl: true,
	}))
}

func (inst *ComputerKeyMapBuilder) ToAction(action *ActionEvent) {
	if inst == nil || action == nil || !inst.hasKey() {
		return
	}

	up := action.FlagKeyUp
	down := action.FlagKeyDown
	if !up && !down {
		up = action.Action == ActionKeyUp
		down = !up
	}
	inst.items = append(inst.items, inst.newMapping(up, down, action))
}

func (inst *ComputerKeyMapBuilder) hasKey() bool {
	if inst == nil {
		return false
	}
	if inst.name != "" {
		return true
	}
	return inst.code > CKeyCodeMin && inst.code < CKeyCodeMax
}

func (inst *ComputerKeyMapBuilder) newMapping(up, down bool, action *ActionEvent) *ComputerKeyMapping {
	return &ComputerKeyMapping{
		FlagKeyUp:   up,
		FlagKeyDown: down,
		Code:        inst.code,
		Name:        inst.name,
		Action:      action,
	}
}

////////////////////////////////////////////////////////////////////////////////
// ComputerKeyMapping

type ComputerKeyMapping struct {

	// flags

	FlagKeyUp   bool
	FlagKeyDown bool

	// key-info

	Code ComputerKeyCode
	Name ComputerKeyName
	URI  ComputerKeyURI

	// action

	Action *ActionEvent
}

func (inst *ComputerKeyMapping) ComputeURI(rewrite bool) ComputerKeyURI {

	if inst == nil {
		return ""
	}

	scheme := ""
	name := inst.Name
	code := inst.Code
	u1 := ""

	if inst.FlagKeyDown {
		scheme = "down"
	}
	if inst.FlagKeyUp {
		scheme = "up"
	}

	if code != 0 {
		u1 = fmt.Sprintf("%s:%s", scheme, code.String())
	} else {
		u1 = fmt.Sprintf("%s:%s", scheme, name.String())
	}

	u2 := ComputerKeyURI(u1)
	if rewrite {
		inst.URI = u2
	}
	return u2
}

////////////////////////////////////////////////////////////////////////////////
// EOF
