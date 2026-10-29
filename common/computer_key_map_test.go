package common

import "testing"

func TestComputerKeyMapFind(t *testing.T) {
	keyMap := new(ComputerKeyMap)
	downA := &ComputerKeyMapping{
		FlagKeyDown: true,
		Code:        CKeyCodeA,
		Action: &ActionEvent{
			Action:   ActionKeyDown,
			FlagCKey: true,
			KeyCode:  CKeyCodeA,
		},
	}
	upEnter := &ComputerKeyMapping{
		FlagKeyUp: true,
		Name:      CKeyNameEnter,
	}
	keyMap.Add(downA)
	keyMap.Add(upEnter)

	cases := []struct {
		name string
		want *ComputerKeyMapping
		have *ComputerKeyMapping
	}{
		{
			name: "down-by-code",
			want: &ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA},
			have: downA,
		},
		{
			name: "up-by-name",
			want: &ComputerKeyMapping{FlagKeyUp: true, Name: CKeyNameEnter},
			have: upEnter,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			have, err := keyMap.Find(c.want)
			if err != nil {
				t.Fatalf("Find() error = %v", err)
			}
			if have != c.have {
				t.Fatalf("Find() = %p, want %p", have, c.have)
			}
			if c.want.URI != "" {
				t.Errorf("Find() rewrote lookup URI to %q", c.want.URI)
			}
		})
	}
}

func TestComputerKeyMapFindDistinguishesDirection(t *testing.T) {
	keyMap := new(ComputerKeyMap)
	downA := &ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA}
	upA := &ComputerKeyMapping{FlagKeyUp: true, Code: CKeyCodeA}
	keyMap.Add(downA)
	keyMap.Add(upA)

	have, err := keyMap.Find(&ComputerKeyMapping{FlagKeyUp: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if have != upA {
		t.Errorf("Find(up) = %p, want %p", have, upA)
	}
}

func TestComputerKeyMapFindReplacement(t *testing.T) {
	keyMap := new(ComputerKeyMap)
	first := &ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA}
	second := &ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA}
	keyMap.Add(first)
	keyMap.Add(second)

	have, err := keyMap.Find(&ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if have != second {
		t.Errorf("Find() = %p, want replacement %p", have, second)
	}
}

func TestComputerKeyMapFindErrors(t *testing.T) {
	keyMap := new(ComputerKeyMap)

	cases := []struct {
		name string
		inst *ComputerKeyMap
		want *ComputerKeyMapping
	}{
		{
			name: "missing",
			inst: keyMap,
			want: &ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA},
		},
		{
			name: "nil-map",
			inst: nil,
			want: &ComputerKeyMapping{Code: CKeyCodeA},
		},
		{
			name: "nil-mapping",
			inst: keyMap,
			want: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			have, err := c.inst.Find(c.want)
			if err == nil {
				t.Fatal("Find() error = nil, want non-nil")
			}
			if have != nil {
				t.Errorf("Find() = %p, want nil", have)
			}
		})
	}
}

func TestComputerKeyMapAddIgnoresNil(t *testing.T) {
	keyMap := new(ComputerKeyMap)
	keyMap.Add(nil)

	var nilMap *ComputerKeyMap
	nilMap.Add(&ComputerKeyMapping{Code: CKeyCodeA})
}

func TestComputerKeyMapBuilderToNote(t *testing.T) {
	builder := new(ComputerKeyMapBuilder)
	if got := builder.MapKeyCode(CKeyCodeA); got != builder {
		t.Fatal("MapKeyCode() did not return the builder")
	}
	builder.ToNote(60)
	keyMap := builder.Build()

	down, err := keyMap.Find(&ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find(down) error = %v", err)
	}
	assertNoteMapping(t, down, CKeyNameA, ActionNoteOn, true, false, 60)

	up, err := keyMap.Find(&ComputerKeyMapping{FlagKeyUp: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find(up) error = %v", err)
	}
	assertNoteMapping(t, up, CKeyNameA, ActionNoteOff, false, true, 60)
}

func assertNoteMapping(t *testing.T, mapping *ComputerKeyMapping, name ComputerKeyName, action Action, noteOn, noteOff bool, note MidiNoteNumber) {
	t.Helper()
	if mapping.Name != name {
		t.Errorf("mapping.Name = %q, want %q", mapping.Name, name)
	}
	if mapping.Action == nil {
		t.Fatal("mapping.Action is nil")
	}
	if mapping.Action.Action != action {
		t.Errorf("mapping.Action.Action = %d, want %d", mapping.Action.Action, action)
	}
	if !mapping.Action.FlagNote {
		t.Error("mapping.Action.FlagNote = false, want true")
	}
	if mapping.Action.FlagNoteOn != noteOn {
		t.Errorf("mapping.Action.FlagNoteOn = %v, want %v", mapping.Action.FlagNoteOn, noteOn)
	}
	if mapping.Action.FlagNoteOff != noteOff {
		t.Errorf("mapping.Action.FlagNoteOff = %v, want %v", mapping.Action.FlagNoteOff, noteOff)
	}
	if mapping.Action.NoteNumber != note {
		t.Errorf("mapping.Action.NoteNumber = %d, want %d", mapping.Action.NoteNumber, note)
	}
	if mapping.Action.NoteVelocity == 0 {
		t.Error("mapping.Action.NoteVelocity = 0, want a playable velocity")
	}
}

func TestComputerKeyMapBuilderToControlByName(t *testing.T) {
	builder := new(ComputerKeyMapBuilder)
	if got := builder.MapKeyName(CKeyNameSpace); got != builder {
		t.Fatal("MapKeyName() did not return the builder")
	}
	builder.ToControl(ActionPlay)
	keyMap := builder.Build()

	mapping, err := keyMap.Find(&ComputerKeyMapping{
		FlagKeyDown: true,
		Name:        CKeyNameSpace,
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if mapping.Name != CKeyNameSpace || mapping.Code != CKeyCodeMin {
		t.Errorf("mapping key = {%d, %q}, want {%d, %q}",
			mapping.Code, mapping.Name, CKeyCodeMin, CKeyNameSpace)
	}
	if mapping.Action == nil {
		t.Fatal("mapping.Action is nil")
	}
	if mapping.Action.Action != ActionPlay {
		t.Errorf("mapping.Action.Action = %d, want %d", mapping.Action.Action, ActionPlay)
	}
	if !mapping.Action.FlagControl {
		t.Error("mapping.Action.FlagControl = false, want true")
	}
}

func TestComputerKeyMapBuilderToAction(t *testing.T) {
	action := &ActionEvent{
		Action:      ActionKeyDown,
		FlagCKey:    true,
		FlagKeyDown: true,
		KeyCode:     CKeyCodeB,
	}

	builder := new(ComputerKeyMapBuilder)
	builder.MapKeyCode(CKeyCodeA)
	builder.ToAction(action)
	keyMap := builder.Build()

	mapping, err := keyMap.Find(&ComputerKeyMapping{FlagKeyDown: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if mapping.Action != action {
		t.Errorf("mapping.Action = %p, want %p", mapping.Action, action)
	}
}

func TestComputerKeyMapBuilderToActionInfersDirection(t *testing.T) {
	action := &ActionEvent{Action: ActionKeyUp, FlagCKey: true, KeyCode: CKeyCodeB}

	builder := new(ComputerKeyMapBuilder)
	builder.MapKeyCode(CKeyCodeA)
	builder.ToAction(action)
	keyMap := builder.Build()

	mapping, err := keyMap.Find(&ComputerKeyMapping{FlagKeyUp: true, Code: CKeyCodeA})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if mapping.Action != action {
		t.Errorf("mapping.Action = %p, want %p", mapping.Action, action)
	}
}

func TestComputerKeyMapBuilderEmpty(t *testing.T) {
	var nilBuilder *ComputerKeyMapBuilder
	if keyMap := nilBuilder.Build(); keyMap == nil {
		t.Fatal("nil builder Build() = nil")
	}
	if got := nilBuilder.MapKeyName(CKeyNameA); got != nil {
		t.Errorf("nil MapKeyName() = %p, want nil", got)
	}
	if got := nilBuilder.MapKeyCode(CKeyCodeA); got != nil {
		t.Errorf("nil MapKeyCode() = %p, want nil", got)
	}
	nilBuilder.ToNote(60)
	nilBuilder.ToControl(ActionPlay)
	nilBuilder.ToAction(&ActionEvent{Action: ActionPlay})

	keyMap := new(ComputerKeyMapBuilder).Build()
	if _, err := keyMap.Find(&ComputerKeyMapping{Code: CKeyCodeA}); err == nil {
		t.Fatal("empty builder produced a mapping")
	}
}
