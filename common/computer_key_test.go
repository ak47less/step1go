package common

import (
	"sort"
	"testing"
)

func TestComputerKeyCodeName(t *testing.T) {
	cases := []struct {
		name string
		code ComputerKeyCode
		want ComputerKeyName
	}{
		{"letter", CKeyCodeA, CKeyNameA},
		{"space", CKeyCodeSpace, CKeyNameSpace},
		{"backspace", CKeyCodeBackspace, CKeyNameBackspace},
		{"numpad-enter", CKeyCodeNumpadEnter, CKeyNameNumpadEnter},
		{"min-sentinel", CKeyCodeMin, CKeyNameMin},
		{"max-sentinel", CKeyCodeMax, CKeyNameMin},
		{"negative", ComputerKeyCode(-1), CKeyNameMin},
		{"unknown", ComputerKeyCode(9999), CKeyNameMin},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.code.Name(); got != c.want {
				t.Errorf("ComputerKeyCode(%d).Name() = %q, want %q", int(c.code), got, c.want)
			}
		})
	}
}

func TestComputerKeyNameCode(t *testing.T) {
	cases := []struct {
		name string
		key  ComputerKeyName
		want ComputerKeyCode
	}{
		{"letter", CKeyNameA, CKeyCodeA},
		{"space", CKeyNameSpace, CKeyCodeSpace},
		{"backspace", CKeyNameBackspace, CKeyCodeBackspace},
		{"numpad-enter", CKeyNameNumpadEnter, CKeyCodeNumpadEnter},
		{"min-sentinel", CKeyNameMin, CKeyCodeMin},
		{"max-sentinel", CKeyNameMax, CKeyCodeMin},
		{"empty", "", CKeyCodeMin},
		{"unknown", "unknown", CKeyCodeMin},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.key.Code(); got != c.want {
				t.Errorf("ComputerKeyName(%q).Code() = %d, want %d", c.key, got, c.want)
			}
		})
	}
}

func TestComputerKeyPrintableCodesMatchASCII(t *testing.T) {
	// 不含空格的 47 个可显示按键。
	printable := "`1234567890-=qwertyuiop[]\\asdfghjkl;'zxcvbnm,./"
	if got, want := len([]rune(printable)), 47; got != want {
		t.Fatalf("printable keyboard test data has %d keys, want %d", got, want)
	}

	for _, ch := range printable {
		code := ComputerKeyCode(ch)
		name := ComputerKeyName(string(ch))
		if got := code.Name(); got != name {
			t.Errorf("ComputerKeyCode(%q).Name() = %q, want %q", ch, got, name)
		}
		if got := name.Code(); got != code {
			t.Errorf("ComputerKeyName(%q).Code() = %d, want %d", name, got, code)
		}
	}

	if CKeyCodeSpace != ComputerKeyCode(' ') {
		t.Errorf("CKeyCodeSpace = %d, want ASCII space (%d)", CKeyCodeSpace, ' ')
	}
	if got := CKeyCodeSpace.Name(); got != CKeyNameSpace {
		t.Errorf("space.Name() = %q, want %q", got, CKeyNameSpace)
	}
	if got := CKeyNameSpace.Code(); got != CKeyCodeSpace {
		t.Errorf("space.Code() = %d, want %d", got, CKeyCodeSpace)
	}
}

func TestComputerKeyboardLists(t *testing.T) {
	keyboard := new(ComputerKeyboard)

	codes := keyboard.ListCodes()
	names := keyboard.ListNames()
	all := keyboard.ListAll()

	const keyCount = 104
	if got := len(codes); got != keyCount {
		t.Fatalf("ListCodes() length = %d, want %d", got, keyCount)
	}
	if got := len(names); got != keyCount {
		t.Fatalf("ListNames() length = %d, want %d", got, keyCount)
	}
	if got := len(all); got != keyCount {
		t.Fatalf("ListAll() length = %d, want %d", got, keyCount)
	}

	seenCodes := make(map[ComputerKeyCode]bool, keyCount)
	seenNames := make(map[ComputerKeyName]bool, keyCount)
	for i, info := range all {
		if info == nil {
			t.Fatalf("ListAll()[%d] is nil", i)
		}
		if info.Code <= CKeyCodeMin || info.Code >= CKeyCodeMax {
			t.Errorf("ListAll()[%d].Code = %d, outside keyboard code range", i, info.Code)
		}
		if seenCodes[info.Code] {
			t.Errorf("duplicate key code %d", info.Code)
		}
		if seenNames[info.Name] {
			t.Errorf("duplicate key name %q", info.Name)
		}
		seenCodes[info.Code] = true
		seenNames[info.Name] = true

		if codes[i] != info.Code || names[i] != info.Name {
			t.Errorf("list order mismatch at %d: code=%d/%d name=%q/%q",
				i, codes[i], info.Code, names[i], info.Name)
		}
		if got := info.Code.Name(); got != info.Name {
			t.Errorf("ComputerKeyCode(%d).Name() = %q, want %q", info.Code, got, info.Name)
		}
		if got := info.Name.Code(); got != info.Code {
			t.Errorf("ComputerKeyName(%q).Code() = %d, want %d", info.Name, got, info.Code)
		}
	}
}

func TestComputerKeyboardListsReturnCopies(t *testing.T) {
	keyboard := new(ComputerKeyboard)

	codes := keyboard.ListCodes()
	names := keyboard.ListNames()
	all := keyboard.ListAll()

	codes[0] = CKeyCodeMin
	names[0] = CKeyNameMax
	all[0].Code = CKeyCodeMin
	all[0].Name = CKeyNameMax

	if got := keyboard.ListCodes()[0]; got != CKeyCodeEscape {
		t.Errorf("ListCodes() shares its backing data: got %d", got)
	}
	if got := keyboard.ListNames()[0]; got != CKeyNameEscape {
		t.Errorf("ListNames() shares its backing data: got %q", got)
	}
	first := keyboard.ListAll()[0]
	if first.Code != CKeyCodeEscape || first.Name != CKeyNameEscape {
		t.Errorf("ListAll() shares its key objects: got {%d, %q}", first.Code, first.Name)
	}
}

func TestComputerKeyString(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"code", CKeyCodeA.String(), "a(97)"},
		{"invalid-code", ComputerKeyCode(9999).String(), "min(9999)"},
		{"name", CKeyNameSpace.String(), "space"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("String() = %q, want %q", c.got, c.want)
			}
		})
	}
}

func TestComputerKeyListAll(t *testing.T) {

	ckb := new(ComputerKeyboard)
	all := ckb.ListAll()

	sort.Slice(all, func(i, j int) bool {
		a := all[i]
		b := all[j]
		return (a.Code < b.Code)
	})

	for idx, item := range all {

		code := item.Code
		name := item.Name

		t.Logf("[ComputerKeyboard index:%d code:%d name:'%s' ]", idx, code, name)
	}

}
