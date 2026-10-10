package step1core

/*
MIDI-Key-Map:

[Note index:127 midi:127 group:9 name:G9 freq:12543.854]
[Note index:126 midi:126 group:9 name:F#9 freq:11839.82]
[Note index:125 midi:125 group:9 name:F9 freq:11175.302]
...
[Note index:108 midi:108 group:8 name:C8  freq:4186] // max(piano)
...
[Note index:71 midi:71 group:4 name:B4 freq:493.88336]
[Note index:70 midi:70 group:4 name:A#4 freq:466.1638]
[Note index:69 midi:69 group:4 name:A4 freq:440.0]
[Note index:68 midi:68 group:4 name:G#4 freq:415.3047]
[Note index:67 midi:67 group:4 name:G4 freq:391.99542]
[Note index:66 midi:66 group:4 name:F#4 freq:369.9944]
[Note index:65 midi:65 group:4 name:F4 freq:349.22818]
[Note index:64 midi:64 group:4 name:E4 freq:329.6275]
[Note index:63 midi:63 group:4 name:D#4 freq:311.12692]
[Note index:62 midi:62 group:4 name:D4 freq:293.6647]
[Note index:61 midi:61 group:4 name:C#4 freq:277.18256]
[Note index:60 midi:60 group:4 name:C4 freq:261.6255]
...
[Note index:21 midi:21 group:0 name:A0  freq:27.5] // min(piano)
...
[Note index:2 midi:2 group:-1 name:D-1 freq:9.177022]
[Note index:1 midi:1 group:-1 name:C#-1 freq:8.661955]
[Note index:0 midi:0 group:-1 name:C-1 freq:8.1757965]

*/

////////////////////////////////////////////////////////////////////////////////

// beats per minute: 节拍速度,  [min,max]=[10,500]
type BPM float32

// 频率,  [min,max]=[20,20000]
type Frequency float32

// 节拍
type Tempo struct {
	Upper uint `json:"upper"`
	Lower uint `json:"lower"`
}

////////////////////////////////////////////////////////////////////////////////
// Tone : 音调
//

// 音调值 (调值), [C(0), C#(1) , D(2) , ... , A(9) , A#(10) , B(11) ]
type ToneNumber int8

const (
	ToneNumberC ToneNumber = iota
	ToneNumberCS
	ToneNumberD
	ToneNumberDS
	ToneNumberE

	ToneNumberF
	ToneNumberFS
	ToneNumberG
	ToneNumberGS
	ToneNumberA
	ToneNumberAS
	ToneNumberB
)

// 音调名称 (调名), 例如 : "C","F#","G#" ...
type ToneName string

const (
	ToneNameC  ToneName = "C"
	ToneNameCS ToneName = "C#"
	ToneNameD  ToneName = "D"
	ToneNameDS ToneName = "D#"
	ToneNameE  ToneName = "E"

	ToneNameF  ToneName = "F"
	ToneNameFS ToneName = "F#"
	ToneNameG  ToneName = "G"
	ToneNameGS ToneName = "G#"
	ToneNameA  ToneName = "A"
	ToneNameAS ToneName = "A#"
	ToneNameB  ToneName = "B"
)

// Tone
type Tone struct {
	Base   *Tone      // 指向 Base-Tone
	Offset int8       // 相对 base 的偏移数值 (this.Number = this.Base.Number + this.Offset)
	Number ToneNumber // 这个 tone 的音调
	Name   ToneName   // 这个 tone 的名称
}

type ToneSet struct {
	base *Tone
	all  []*Tone
}

type ToneSetBuilder struct {
	ts ToneSet
}

/*************************************/
/* impl: Tone                        */
/*************************************/

func (inst *Tone) String() string {
	if inst == nil {
		return ""
	}
	return inst.Name.String()
}

func (inst *Tone) Equals(other *Tone) bool {
	if inst == nil || other == nil {
		return false
	}
	n1 := inst.Number.Normalize()
	n2 := other.Number.Normalize()
	return n1 == n2
}

/*************************************/
/* impl: ToneNumber                  */
/*************************************/

func (num ToneNumber) String() string {
	name := num.Name()
	return name.String()
}

var theToneNumber12Tones = [12]ToneName{

	ToneNameC,
	ToneNameCS,
	ToneNameD,
	ToneNameDS,
	ToneNameE,

	ToneNameF,
	ToneNameFS,
	ToneNameG,
	ToneNameGS,
	ToneNameA,
	ToneNameAS,
	ToneNameB,
}

func (num ToneNumber) Name() ToneName {
	n2 := num.Normalize()
	return theToneNumber12Tones[int(n2)]
}

func (num ToneNumber) Normalize() ToneNumber {
	for num < 0 {
		num += 12
	}
	return (num % 12)
}

/*************************************/
/* impl: ToneName                    */
/*************************************/

func (name ToneName) String() string {
	return string(name)
}

func (name ToneName) Number() ToneNumber {

	const (
		min = 'C' // the min rune
		max = 'B' // the max rune
		na  = 0   // value(N/A)
	)

	size := len(name)
	var ch0, ch1 byte
	var num ToneNumber

	switch size {
	case 1:
		ch0 = name[0]
		ch1 = '.'
	case 2:
		ch0 = name[0]
		ch1 = name[1]
	default:
		return na
	}

	// to UpperCase
	if 'a' <= ch0 && ch0 <= 'z' {
		ch0 = ch0 - 'a' + 'A'
	}

	// for ch0:
	// 0..1...2..3...4........5..6...7..8........9..10..11
	// C..C#..D..D#..E........F..F#..G..G#.......A..A#..B.

	switch ch0 {
	case 'C':
		num = 0
	case 'D':
		num = 2
	case 'E':
		num = 4
	case 'F':
		num = 5
	case 'G':
		num = 7
	case 'A':
		num = 9
	case 'B':
		num = 11
	default:
		return na
	}

	// for '#'
	switch ch1 {
	case '#':
		num++
	case 'b':
		num--
	}
	return num
}

/*************************************/
/* impl: ToneSet                     */
/*************************************/

func (inst *ToneSet) Reset() *ToneSet {
	if inst != nil {
		inst.all = nil
		inst.base = nil
	}
	return inst
}

func (inst *ToneSet) Add(item *Tone) *ToneSet {
	if inst != nil && item != nil {
		inst.all = append(inst.all, item)
		if inst.base == nil {
			inst.base = item
		}
	}
	return inst
}

func (inst *ToneSet) Normalize() *ToneSet {

	// todo

	return inst
}

func (inst *ToneSet) Rebase(base2 *Tone) *ToneSet {

	ts2 := new(ToneSet)
	ts1 := inst
	if ts1 == nil || base2 == nil {
		return ts2
	}

	base1 := ts1.base
	if base1 == nil {
		return ts2
	}

	// check base eq
	if base2.Equals(base1) {
		*ts2 = *ts1
		return ts2
	}

	// copy items  (n2 = n1 + shift )

	base2copy := new(Tone)
	*base2copy = *base2

	base2 = base2copy
	item2 := base2copy
	shift := base2.Number - base1.Number

	for _, item1 := range ts1.all {

		if item1.Equals(base1) {
			// base
			item2 = base2copy
			ts2.base = base2copy
		} else {
			// other
			item2 = new(Tone)
		}

		num := item1.Number + shift
		item2.Number = num
		item2.Name = num.Name()
		item2.Base = base2
		item2.Offset = int8(num - base2.Number)

		ts2.Add(item2)
	}

	return ts2
}

/*************************************/
/* impl: ToneSetBuilder              */
/*************************************/

// dst is optional (can be nil)
func (inst *ToneSetBuilder) Build(dst *ToneSet) *ToneSet {
	if dst == nil {
		dst = new(ToneSet)
	}
	if inst != nil {
		*dst = inst.ts
	}
	return dst
}

// pattern 是一个表示模式的字符串, 例如: "0...1..2", 表示一个大三和弦
func (inst *ToneSetBuilder) SetPattern(pattern string) *ToneSetBuilder {

	const (
		dot   = '.'
		space = ' '
	)

	buf := []rune(pattern)
	index := 0
	root := new(Tone)
	t2 := root
	c0 := ToneNumberC

	root.Offset = 0
	root.Base = root
	root.Number = c0
	root.Name = c0.Name()

	for _, ch := range buf {

		if ch == dot {
			index++
			continue
		} else if '1' <= ch && ch <= '9' {
		} else if 'a' <= ch && ch <= 'z' {
		} else if 'A' <= ch && ch <= 'Z' {
		} else {
			continue
		}

		if index > 0 {
			t2 = new(Tone)
			off := index
			num := root.Number + ToneNumber(off)
			t2.Offset = int8(off)
			t2.Base = root
			t2.Number = num
			t2.Name = num.Name()
		} else {
			t2 = root
		}

		inst.ts.Add(t2)
		index++
	}

	return inst
}

func (inst *ToneSetBuilder) Reset() *ToneSetBuilder {
	inst.ts.Reset()
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// Pitch : 音高

// PitchNumber : 音高值, aka.MIDI 键值
type PitchNumber uint

// PitchName : 音名, 例如:'C#5', 'F3', ...
type PitchName string

// 音程
type PitchInterval int8

type PitchGroup int8

type PitchInfo struct {
	Number    PitchNumber
	Name      PitchName
	Group     PitchGroup
	Frequency Frequency
}

////////////////////////////////////////////////////////////////////////////////
// Chord : 和弦

type ChordMode string

const (
	ChordModeMock ChordMode = "mock"

	// 3

	ChordModeMaj  ChordMode = "M"
	ChordModeMin  ChordMode = "m"
	ChordModeAug  ChordMode = "aug"
	ChordModeDim  ChordMode = "dim"
	ChordModeSus2 ChordMode = "sus2"
	ChordModeSus4 ChordMode = "sus4"
	ChordMode5    ChordMode = "5"
	ChordModeM6   ChordMode = "M6"
	ChordModeMin6 ChordMode = "min6"
	ChordMode6p9  ChordMode = "6/9"
	ChordModeM6p9 ChordMode = "m6/9"

	// 7

	ChordModeMaj7  ChordMode = "M7"
	ChordModeMin7  ChordMode = "m7"
	ChordMode7     ChordMode = "7"
	ChordModeM7b5  ChordMode = "m7b5"
	ChordModeDim7  ChordMode = "dim7"
	ChordModePM7   ChordMode = "+M7"
	ChordModeP7    ChordMode = "+7"
	ChordMode7sus4 ChordMode = "7sus4"
	ChordMode7b5   ChordMode = "7b5"

	// 9

	ChordMode9    ChordMode = "9"
	ChordModeMaj9 ChordMode = "maj9"
	ChordModeMin9 ChordMode = "m9"
	ChordModeM9s5 ChordMode = "9#5"

	// 11

	ChordMode11    ChordMode = "11"
	ChordModeMin11 ChordMode = "m11"
	ChordModeMaj11 ChordMode = "maj11"

	// 13

)

type ChordTemplate struct {
	ToneSet

	Mode ChordMode
}

type ChordInstance struct {
	Template ChordTemplate // 模板
	Root     Tone          // 根音
}

////////////////////////////////////////////////////////////////////////////////
// Scale : 调式

type ScaleMode string

const (
	ScaleModeMock ScaleMode = "mock"

	// 大调

	ScaleModeMajor        ScaleMode = "major"         // 大调
	ScaleModeNaturalMajor ScaleMode = "natural_major" // 自然大调

	// 小调

	ScaleModeMinor         ScaleMode = "minor"          // 小调
	ScaleModeNaturalMinor  ScaleMode = "natural_minor"  // 自然小调
	ScaleModeHarmonicMinor ScaleMode = "harmonic_minor" // 和声小调
	ScaleModeMelodicMinor  ScaleMode = "melodic_minor"  // 旋律小调

	// 五声音阶

	ScaleModePentatonic      ScaleMode = "Pentatonic"      // 五声音阶
	ScaleModeMajorPentatonic ScaleMode = "MajorPentatonic" // 大调五声音阶
	ScaleModeMinorPentatonic ScaleMode = "MinorPentatonic" // 小调五声音阶

)

type ScaleTemplate struct {
	ToneSet

	Mode ScaleMode
}

type ScaleInstance struct {
	Template ScaleTemplate // 模板
	Primary  Tone          // 基调 (主音)
}

////////////////////////////////////////////////////////////////////////////////
// note

// MIDI 按键编码, 例如:(value=69,A4,440Hz), (value=60,C4(middle_C),261.6Hz)
type NotePitch = PitchNumber

////////////////////////////////////////////////////////////////////////////////
// EOF
