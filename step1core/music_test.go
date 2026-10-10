package step1core

import "testing"

////////////////////////////////////////////////////////////////////////////////
// 测试辅助

// mockTone 构造一个音调值为 num 的 Tone (名称由 num 归一化得到)
func mockTone(num ToneNumber) *Tone {
	inst := new(Tone)
	inst.Number = num
	inst.Name = num.Name()
	return inst
}

// mockTriad 构造一个 C 大三和弦 (C,E,G) 的 ToneSet
func mockTriad() *ToneSet {
	ts := new(ToneSet)
	ts.Add(mockTone(ToneNumberC))
	ts.Add(mockTone(ToneNumberE))
	ts.Add(mockTone(ToneNumberG))
	return ts
}

////////////////////////////////////////////////////////////////////////////////
// Tone

func TestToneString(t *testing.T) {

	cases := []struct {
		name string
		inst *Tone
		want string
	}{
		{name: "C", inst: &Tone{Name: ToneNameC}, want: "C"},
		{name: "F#", inst: &Tone{Name: ToneNameFS}, want: "F#"},
		{name: "empty-name", inst: new(Tone), want: ""},
		{name: "nil", inst: nil, want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.inst.String()
			if got != c.want {
				t.Errorf("Tone.String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestToneEquals(t *testing.T) {

	var nilTone *Tone

	cases := []struct {
		name string
		a    *Tone
		b    *Tone
		want bool
	}{
		{"same-number", &Tone{Number: ToneNumberC}, &Tone{Number: ToneNumberC}, true},
		{"same-number-diff-name", &Tone{Number: ToneNumberD, Name: ToneNameD}, &Tone{Number: ToneNumberD, Name: ToneNameDS}, true},
		{"octave-eq", &Tone{Number: ToneNumberC}, &Tone{Number: ToneNumber(12)}, true},
		{"negative-eq", &Tone{Number: ToneNumber(-1)}, &Tone{Number: ToneNumber(11)}, true},
		{"diff-number", &Tone{Number: ToneNumberC}, &Tone{Number: ToneNumberD}, false},
		{"this-nil", nilTone, &Tone{Number: ToneNumberC}, false},
		{"other-nil", &Tone{Number: ToneNumberC}, nilTone, false},
		{"both-nil", nilTone, nilTone, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.a.Equals(c.b)
			if got != c.want {
				t.Errorf("Tone.Equals() = %v, want %v", got, c.want)
			}
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// ToneNumber

func TestToneNumberString(t *testing.T) {

	cases := []struct {
		name string
		num  ToneNumber
		want string
	}{
		{"C", ToneNumberC, "C"},
		{"C#", ToneNumberCS, "C#"},
		{"D#", ToneNumberDS, "D#"},
		{"A#", ToneNumberAS, "A#"},
		{"B", ToneNumberB, "B"},
		{"12->C", ToneNumber(12), "C"},
		{"-1->B", ToneNumber(-1), "B"},
		{"-12->C", ToneNumber(-12), "C"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.num.String()
			if got != c.want {
				t.Errorf("ToneNumber(%d).String() = %q, want %q", int(c.num), got, c.want)
			}
		})
	}
}

func TestToneNumberName(t *testing.T) {

	cases := []struct {
		name string
		num  ToneNumber
		want ToneName
	}{
		{"0", ToneNumber(0), ToneNameC},
		{"1", ToneNumber(1), ToneNameCS},
		{"6", ToneNumber(6), ToneNameFS},
		{"11", ToneNumber(11), ToneNameB},
		{"12", ToneNumber(12), ToneNameC},
		{"25", ToneNumber(25), ToneNameCS},
		{"-1", ToneNumber(-1), ToneNameB},
		{"-13", ToneNumber(-13), ToneNameB},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.num.Name()
			if got != c.want {
				t.Errorf("ToneNumber(%d).Name() = %q, want %q", int(c.num), got, c.want)
			}
		})
	}
}

func TestToneNumberNormalize(t *testing.T) {

	cases := []struct {
		name string
		num  ToneNumber
		want ToneNumber
	}{
		{"0", 0, 0},
		{"11", 11, 11},
		{"12", 12, 0},
		{"13", 13, 1},
		{"-1", -1, 11},
		{"-12", -12, 0},
		{"-13", -13, 11},
		{"-25", -25, 11},
		{"100", 100, 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.num.Normalize()
			if got != c.want {
				t.Errorf("ToneNumber(%d).Normalize() = %d, want %d", int(c.num), int(got), int(c.want))
			}
		})
	}
}

// 归一化后的 ToneNumber 必须是合法的 [0,11] 区间
func TestToneNumberNormalizeRange(t *testing.T) {
	for i := -128; i <= 127; i++ {
		got := ToneNumber(i).Normalize()
		if got < 0 || got > 11 {
			t.Fatalf("ToneNumber(%d).Normalize() = %d, out of [0,11]", i, int(got))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// ToneName

func TestToneNameString(t *testing.T) {

	cases := []struct {
		name string
		inst ToneName
		want string
	}{
		{"C", ToneNameC, "C"},
		{"CS", ToneNameCS, "C#"},
		{"FS", ToneNameFS, "F#"},
		{"empty", ToneName(""), ""},
		{"mock", ToneName("mock"), "mock"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.inst.String()
			if got != c.want {
				t.Errorf("ToneName(%q).String() = %q, want %q", string(c.inst), got, c.want)
			}
		})
	}
}

func TestToneNameNumber(t *testing.T) {

	cases := []struct {
		name string
		inst ToneName
		want ToneNumber
	}{
		// 自然音
		{"C", ToneNameC, 0},
		{"D", ToneNameD, 2},
		{"E", ToneNameE, 4},
		{"F", ToneNameF, 5},
		{"G", ToneNameG, 7},
		{"A", ToneNameA, 9},
		{"B", ToneNameB, 11},

		// 升号
		{"C#", ToneNameCS, 1},
		{"D#", ToneNameDS, 3},
		{"F#", ToneNameFS, 6},
		{"G#", ToneNameGS, 8},
		{"A#", ToneNameAS, 10},

		// 降号 (b)
		{"Db", ToneName("Db"), 1},
		{"Eb", ToneName("Eb"), 3},
		{"Gb", ToneName("Gb"), 6},
		{"Ab", ToneName("Ab"), 8},
		{"Bb", ToneName("Bb"), 10},
		{"Cb", ToneName("Cb"), -1},

		// 小写 (会被转换为大写)
		{"c", ToneName("c"), 0},
		{"d#", ToneName("d#"), 3},
		{"b", ToneName("b"), 11},

		// 非法/越界
		{"empty", ToneName(""), 0},
		{"unknown-letter", ToneName("X"), 0},
		{"unknown-tail", ToneName("Cx"), 0},
		{"too-long", ToneName("CCC"), 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.inst.Number()
			if got != c.want {
				t.Errorf("ToneName(%q).Number() = %d, want %d", string(c.inst), int(got), int(c.want))
			}
		})
	}
}

// 12 个音名做一次 Number -> Name 往返, 应当回到自身
func TestToneNameNumberRoundTrip(t *testing.T) {

	all := []ToneName{
		ToneNameC, ToneNameCS, ToneNameD, ToneNameDS, ToneNameE, ToneNameF,
		ToneNameFS, ToneNameG, ToneNameGS, ToneNameA, ToneNameAS, ToneNameB,
	}

	if len(all) != 12 {
		t.Fatalf("expect 12 tone names, got %d", len(all))
	}

	for _, name := range all {
		got := name.Number().Name()
		if got != name {
			t.Errorf("ToneName(%q) round-trip = %q", string(name), string(got))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// ToneSet

func TestToneSetReset(t *testing.T) {

	t.Run("active-set", func(t *testing.T) {
		ts := mockTriad()
		got := ts.Reset()
		if got != ts {
			t.Errorf("ToneSet.Reset() should return itself")
		}
		if ts.base != nil {
			t.Errorf("base should be nil after Reset()")
		}
		if ts.all != nil {
			t.Errorf("all should be nil after Reset()")
		}
	})

	t.Run("empty-set", func(t *testing.T) {
		ts := new(ToneSet)
		if got := ts.Reset(); got != ts {
			t.Errorf("ToneSet.Reset() should return itself")
		}
	})

	t.Run("nil-receiver", func(t *testing.T) {
		var ts *ToneSet
		if got := ts.Reset(); got != nil {
			t.Errorf("(*ToneSet)(nil).Reset() = %v, want nil", got)
		}
	})
}

func TestToneSetAdd(t *testing.T) {

	t.Run("first-item-becomes-base", func(t *testing.T) {
		ts := new(ToneSet)
		t1 := mockTone(ToneNumberC)
		t2 := mockTone(ToneNumberE)
		t3 := mockTone(ToneNumberG)

		if got := ts.Add(t1); got != ts {
			t.Errorf("ToneSet.Add() should return itself")
		}
		ts.Add(t2)
		ts.Add(t3)

		if len(ts.all) != 3 {
			t.Fatalf("len(all) = %d, want 3", len(ts.all))
		}
		if ts.base != t1 {
			t.Errorf("base = %v, want the first added item", ts.base)
		}
		for i, want := range []*Tone{t1, t2, t3} {
			if ts.all[i] != want {
				t.Errorf("all[%d] = %v, want %v", i, ts.all[i], want)
			}
		}
	})

	t.Run("nil-item-ignored", func(t *testing.T) {
		ts := new(ToneSet)
		if got := ts.Add(nil); got != ts {
			t.Errorf("ToneSet.Add(nil) should return itself")
		}
		if ts.base != nil || len(ts.all) != 0 {
			t.Errorf("Add(nil) should not change the set")
		}
	})

	t.Run("nil-receiver", func(t *testing.T) {
		var ts *ToneSet
		if got := ts.Add(mockTone(ToneNumberC)); got != nil {
			t.Errorf("(*ToneSet)(nil).Add() = %v, want nil", got)
		}
	})
}

// 注意: ToneSet.Normalize() 目前是占位实现 (源码中标记为 todo),
// 它会原样返回接收者且不做任何改动。这里先锁定当前行为,
// 待真正实现后, 应把它替换成对归一结果的断言。
func TestToneSetNormalize(t *testing.T) {

	ts := new(ToneSet)
	ts.Add(mockTone(ToneNumberC))
	ts.Add(mockTone(ToneNumber(14)))
	ts.Add(mockTone(ToneNumber(-1)))

	got := ts.Normalize()

	if got != ts {
		t.Fatalf("ToneSet.Normalize() = %v, want the receiver itself", got)
	}
	if got.base != ts.base {
		t.Errorf("Normalize() should keep base")
	}
	if len(got.all) != 3 {
		t.Fatalf("len(all) = %d, want 3", len(got.all))
	}
	if got.all[1].Number != ToneNumber(14) {
		t.Errorf("Normalize() (todo-stub) should not modify items, got %d", int(got.all[1].Number))
	}
}

func TestToneSetRebase(t *testing.T) {

	t.Run("same-base-shares-data", func(t *testing.T) {

		ts1 := mockTriad()
		base2 := mockTone(ToneNumberC) // 与 ts1.base 等音

		ts2 := ts1.Rebase(base2)
		if ts2 == nil {
			t.Fatalf("ToneSet.Rebase() = nil")
		}
		if ts2 == ts1 {
			t.Errorf("Rebase() should return a new ToneSet")
		}
		if ts2.base != ts1.base {
			t.Errorf("same base: base pointer should be shared")
		}
		if len(ts2.all) != len(ts1.all) {
			t.Fatalf("len(all) = %d, want %d", len(ts2.all), len(ts1.all))
		}
		if &ts2.all[0] != &ts1.all[0] {
			t.Errorf("same base: items should share the same backing array")
		}
	})

	t.Run("rebase-C-to-D", func(t *testing.T) {

		ts1 := mockTriad() // C, E, G
		base2 := mockTone(ToneNumberD)

		ts2 := ts1.Rebase(base2)
		if len(ts2.all) != 3 {
			t.Fatalf("len(all) = %d, want 3", len(ts2.all))
		}
		if ts2.base != ts2.all[0] {
			t.Errorf("new base should be the first item")
		}
		if ts2.base.Number != ToneNumberD || ts2.base.Name != ToneNameD {
			t.Errorf("base = (%d,%q), want (2,\"D\")", int(ts2.base.Number), string(ts2.base.Name))
		}
		if ts2.base.Offset != 0 {
			t.Errorf("base.Offset = %d, want 0", int(ts2.base.Offset))
		}

		wantNum := []ToneNumber{ToneNumberD, ToneNumber(6) /*F#*/, ToneNumberA}
		wantOff := []int8{0, 4, 7}

		for i, item := range ts2.all {
			if item.Number != wantNum[i] {
				t.Errorf("all[%d].Number = %d, want %d", i, int(item.Number), int(wantNum[i]))
			}
			if item.Name != wantNum[i].Name() {
				t.Errorf("all[%d].Name = %q, want %q", i, string(item.Name), string(wantNum[i].Name()))
			}
			if item.Offset != wantOff[i] {
				t.Errorf("all[%d].Offset = %d, want %d", i, int(item.Offset), int(wantOff[i]))
			}
			if item.Base != ts2.base {
				t.Errorf("all[%d].Base should point to the new base", i)
			}
		}

		// 源集合不应被改动
		if ts1.base.Number != ToneNumberC || len(ts1.all) != 3 || ts1.all[1].Number != ToneNumberE {
			t.Errorf("Rebase() should not modify the source ToneSet")
		}
		// 新旧集合不共享元素
		if &ts2.all[0] == &ts1.all[0] {
			t.Errorf("rebased items should be copies")
		}
	})

	t.Run("nil-base", func(t *testing.T) {

		ts1 := mockTriad()
		ts2 := ts1.Rebase(nil)

		if ts2 == nil {
			t.Fatalf("ToneSet.Rebase(nil) = nil")
		}
		if ts2.base != nil || len(ts2.all) != 0 {
			t.Errorf("Rebase(nil) should return an empty ToneSet")
		}
	})

	t.Run("empty-set", func(t *testing.T) {

		ts1 := new(ToneSet)
		ts2 := ts1.Rebase(mockTone(ToneNumberC))

		if ts2 == nil {
			t.Fatalf("empty ToneSet.Rebase() = nil")
		}
		if ts2.base != nil || len(ts2.all) != 0 {
			t.Errorf("empty ToneSet.Rebase() should return an empty ToneSet")
		}
	})

	t.Run("nil-receiver", func(t *testing.T) {

		var ts1 *ToneSet
		ts2 := ts1.Rebase(mockTone(ToneNumberC))

		if ts2 == nil {
			t.Fatalf("(*ToneSet)(nil).Rebase() = nil")
		}
		if ts2.base != nil || len(ts2.all) != 0 {
			t.Errorf("(*ToneSet)(nil).Rebase() should return an empty ToneSet")
		}
	})
}

////////////////////////////////////////////////////////////////////////////////
// ToneSetBuilder

func TestToneSetBuilderBuild(t *testing.T) {

	t.Run("nil-builder-nil-dst", func(t *testing.T) {
		var b *ToneSetBuilder
		ts := b.Build(nil)
		if ts == nil {
			t.Fatalf("(*ToneSetBuilder)(nil).Build(nil) = nil")
		}
		if ts.base != nil || len(ts.all) != 0 {
			t.Errorf("Build() should return an empty ToneSet")
		}
	})

	t.Run("nil-builder-keeps-dst", func(t *testing.T) {
		var b *ToneSetBuilder
		dst := mockTriad()
		got := b.Build(dst)
		if got != dst {
			t.Errorf("Build(dst) should return dst")
		}
		if len(got.all) != 3 {
			t.Errorf("Build(dst) with nil builder should keep dst, len(all) = %d", len(got.all))
		}
	})

	t.Run("copy-into-dst", func(t *testing.T) {
		b := new(ToneSetBuilder)
		b.SetPattern("1...2..3")

		dst := new(ToneSet)
		got := b.Build(dst)
		if got != dst {
			t.Errorf("Build(dst) should return dst")
		}
		if len(got.all) != 3 {
			t.Fatalf("len(all) = %d, want 3", len(got.all))
		}
		if got.base != got.all[0] {
			t.Errorf("base should be the root item")
		}
	})

	t.Run("fresh-dst", func(t *testing.T) {
		b := new(ToneSetBuilder)
		b.SetPattern("1...2..3")
		ts := b.Build(nil)
		if got := b.Build(nil); got == ts {
			t.Errorf("Build(nil) should allocate a new ToneSet each time")
		}
	})
}

func TestToneSetBuilderSetPattern(t *testing.T) {

	t.Run("major-triad", func(t *testing.T) {

		// "1...2..3" : index 0,4,7 -> C, E, G
		b := new(ToneSetBuilder)
		ts := b.SetPattern("1...2..3").Build(nil)

		if len(ts.all) != 3 {
			t.Fatalf("len(all) = %d, want 3", len(ts.all))
		}
		if ts.base != ts.all[0] {
			t.Errorf("base should be the first (root) item")
		}

		wantNum := []ToneNumber{ToneNumberC, ToneNumberE, ToneNumberG}
		wantOff := []int8{0, 4, 7}

		for i, item := range ts.all {
			if item.Number != wantNum[i] {
				t.Errorf("all[%d].Number = %d, want %d", i, int(item.Number), int(wantNum[i]))
			}
			if item.Name != wantNum[i].Name() {
				t.Errorf("all[%d].Name = %q, want %q", i, string(item.Name), string(wantNum[i].Name()))
			}
			if item.Offset != wantOff[i] {
				t.Errorf("all[%d].Offset = %d, want %d", i, int(item.Offset), int(wantOff[i]))
			}
			if item.Base != ts.base {
				t.Errorf("all[%d].Base should point to the root item", i)
			}
		}

		// root 的 Base 指向自身
		if ts.base.Base != ts.base {
			t.Errorf("root.Base should point to itself")
		}
	})

	t.Run("letters-are-triggers", func(t *testing.T) {

		// 字母与数字一样, 只作为触发器; 决定音高的是它的下标
		b := new(ToneSetBuilder)
		ts := b.SetPattern("a...b..c").Build(nil)

		if len(ts.all) != 3 {
			t.Fatalf("len(all) = %d, want 3", len(ts.all))
		}
		wantNum := []ToneNumber{ToneNumberC, ToneNumberE, ToneNumberG}
		for i, item := range ts.all {
			if item.Number != wantNum[i] {
				t.Errorf("all[%d].Number = %d, want %d", i, int(item.Number), int(wantNum[i]))
			}
		}
	})

	t.Run("blank-and-symbol-ignored", func(t *testing.T) {

		// 空格/符号会被忽略, 且不占用下标
		b := new(ToneSetBuilder)
		ts := b.SetPattern("1 - # @ ... 2").Build(nil)

		if len(ts.all) != 2 {
			t.Fatalf("len(all) = %d, want 2", len(ts.all))
		}
		if ts.all[0].Number != ToneNumberC || ts.all[1].Number != ToneNumberE {
			t.Errorf("numbers = [%d,%d], want [0,4]", int(ts.all[0].Number), int(ts.all[1].Number))
		}
	})

	t.Run("empty-pattern", func(t *testing.T) {
		b := new(ToneSetBuilder)
		ts := b.SetPattern("").Build(nil)
		if ts.base != nil || len(ts.all) != 0 {
			t.Errorf("empty pattern should build an empty ToneSet")
		}
	})

	// 锁定当前行为: '0' 不在触发字符范围内 (源码只接受 '1'..'9'/字母),
	// 所以文档注释里的 "0...1..2" 并不会产生注释所说的大三和弦。
	t.Run("doc-example-behavior", func(t *testing.T) {

		b := new(ToneSetBuilder)
		ts := b.SetPattern("0...1..2").Build(nil)

		if len(ts.all) != 2 {
			t.Fatalf("len(all) = %d, want 2 ('0' is skipped)", len(ts.all))
		}
		if ts.all[0].Number != ToneNumber(3) {
			t.Errorf("all[0].Number = %d, want 3", int(ts.all[0].Number))
		}
		if ts.all[1].Number != ToneNumber(6) {
			t.Errorf("all[1].Number = %d, want 6", int(ts.all[1].Number))
		}
	})

	t.Run("accumulate-without-reset", func(t *testing.T) {

		// SetPattern 不清空已有内容, 多次调用会累加
		b := new(ToneSetBuilder)
		b.SetPattern("1...2..3")
		ts := b.SetPattern("1...2..3").Build(nil)

		if len(ts.all) != 6 {
			t.Errorf("len(all) = %d, want 6 (accumulated)", len(ts.all))
		}
	})
}

func TestToneSetBuilderReset(t *testing.T) {

	b := new(ToneSetBuilder)
	ts1 := b.SetPattern("1...2..3").Build(nil)
	if len(ts1.all) != 3 {
		t.Fatalf("len(all) = %d, want 3", len(ts1.all))
	}

	if got := b.Reset(); got != b {
		t.Errorf("ToneSetBuilder.Reset() should return itself")
	}

	ts2 := b.Build(nil)
	if ts2.base != nil || len(ts2.all) != 0 {
		t.Errorf("builder should be empty after Reset()")
	}

	// Reset 之后可以重新构建
	ts3 := b.SetPattern("1...2..3").Build(nil)
	if len(ts3.all) != 3 {
		t.Errorf("len(all) = %d, want 3 after rebuild", len(ts3.all))
	}

	// 之前 Build 出来的集合不受影响
	if len(ts1.all) != 3 {
		t.Errorf("previously built ToneSet should not be affected by Reset()")
	}
}

////////////////////////////////////////////////////////////////////////////////
// EOF
