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
}

func (inst *ComputerKeyMapBuilder) Build() *ComputerKeyMap {

	// todo

}

func (inst *ComputerKeyMapBuilder) MapKeyName(name ComputerKeyName) *ComputerKeyMapBuilder {

	// todo

}

func (inst *ComputerKeyMapBuilder) MapKeyCode(code ComputerKeyCode) *ComputerKeyMapBuilder {

	// todo

}

func (inst *ComputerKeyMapBuilder) ToNote(note MidiNoteNumber) {

	// todo

}

func (inst *ComputerKeyMapBuilder) ToControl(action Action) {

	// todo

}

func (inst *ComputerKeyMapBuilder) ToAction(action *ActionEvent) {

	// todo

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
