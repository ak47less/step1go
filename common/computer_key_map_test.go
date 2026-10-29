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
