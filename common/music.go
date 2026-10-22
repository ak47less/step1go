package common

import (
	"math"
	"strconv"
)

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
	ToneNameC  ToneName = "C"  // 0
	ToneNameCS ToneName = "C#" // 1
	ToneNameD  ToneName = "D"  // 2
	ToneNameDS ToneName = "D#" // 3
	ToneNameE  ToneName = "E"  // 4

	ToneNameF  ToneName = "F"  // 5
	ToneNameFS ToneName = "F#" // 6
	ToneNameG  ToneName = "G"  // 7
	ToneNameGS ToneName = "G#" // 8
	ToneNameA  ToneName = "A"  // 9
	ToneNameAS ToneName = "A#" // 10
	ToneNameB  ToneName = "B"  // 11
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

// Normalize : 整理集合, 使其中各个 Tone 与 base 之间的关系保持一致, 即满足 :
//
//	base  : 必定存在 (缺省时取第一个有效项), 并且 Base 指向自身, Offset = 0
//	其他项: Base 指向 base, Offset = Number - base.Number, Name = Number.Name()
//
// 说明 :
//   - nil 项会被丢弃; base 若不在集合中, 会被插入到最前面; 其余项保持原有顺序。
//   - Number 保持原值 (它同时表示音级和相对 base 的音程, 例如 9th = 14),
//     只重置由它派生的 Name / Offset / Base, 因此本方法可重复调用 (幂等)。
//   - 接收者为 nil 时返回一个新建的空集合 (与 Timeline.Normalize 一致)。
func (inst *ToneSet) Normalize() *ToneSet {

	if inst == nil {
		return new(ToneSet)
	}

	// 1. 收集有效的项 (丢弃 nil, 保持顺序)
	items := make([]*Tone, 0, len(inst.all))
	for _, item := range inst.all {
		if item != nil {
			items = append(items, item)
		}
	}

	// 2. 确定 base : 优先使用现有的 base, 否则取第一个有效项
	base := inst.base
	if base == nil && len(items) > 0 {
		base = items[0]
	}

	// base 不在集合中时, 插入到最前面
	if base != nil {
		found := false
		for _, item := range items {
			if item == base {
				found = true
				break
			}
		}
		if !found {
			items = append([]*Tone{base}, items...)
		}
	}

	// 3. 空集合
	if base == nil {
		inst.base = nil
		inst.all = nil
		return inst
	}

	// 4. 修复 base 自身
	base.Base = base
	base.Name = base.Number.Name()
	base.Offset = 0

	// 5. 修复其余各项 : 相对 base 重新计算
	for _, item := range items {
		if item == base {
			continue
		}
		item.Base = base
		item.Name = item.Number.Name()
		item.Offset = int8(int(item.Number) - int(base.Number))
	}

	inst.base = base
	inst.all = items
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
		} else if '0' <= ch && ch <= '9' {
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
type PitchNumber = MidiNoteNumber

// PitchName : 音名, 例如:'C#5', 'F3', ...
type PitchName string

// 音程
type PitchInterval int8

type PitchGroup int8

type PitchInfo struct {
	Number    PitchNumber
	Name      PitchName
	Group     PitchGroup
	Tone      ToneNumber
	Frequency Frequency
}

func (num PitchNumber) GetInfo(dst *PitchInfo) *PitchInfo {

	if dst == nil {
		dst = new(PitchInfo)
	}

	group := int(num)/12 - 1
	tone_no := ToneNumber(num % 12)
	tone_name := tone_no.Name()

	dst.Number = num
	dst.Name = PitchName(tone_name.String() + strconv.Itoa(group))
	dst.Group = PitchGroup(group)
	dst.Frequency = Frequency(440 * math.Pow(2, (float64(num)-69)/12))
	dst.Tone = tone_no

	return dst
}

////////////////////////////////////////////////////////////////////////////////
// Chord : 和弦

type ChordMode string

const (
	ChordModeMock ChordMode = "mock" // 模拟和弦（占位）

	// 三和弦与六和弦

	ChordModeMaj     ChordMode = "M"     // 大三和弦
	ChordModeMin     ChordMode = "m"     // 小三和弦
	ChordModeAug     ChordMode = "aug"   // 增三和弦
	ChordModeDim     ChordMode = "dim"   // 减三和弦
	ChordModeSus2    ChordMode = "sus2"  // 挂二和弦
	ChordModeSus4    ChordMode = "sus4"  // 挂四和弦
	ChordMode5       ChordMode = "5"     // 强力和弦（纯五度）
	ChordModeM6      ChordMode = "M6"    // 大六和弦
	ChordModeMin6    ChordMode = "min6"  // 小六和弦
	ChordMode6p9     ChordMode = "6/9"   // 六九和弦
	ChordModeM6p9    ChordMode = "m6/9"  // 小六九和弦
	ChordModeAdd9    ChordMode = "add9"  // 加九和弦
	ChordModeMinAdd9 ChordMode = "madd9" // 小加九和弦

	// 七和弦

	ChordModeMaj7    ChordMode = "M7"    // 大七和弦
	ChordModeMin7    ChordMode = "m7"    // 小七和弦
	ChordMode7       ChordMode = "7"     // 属七和弦
	ChordModeM7b5    ChordMode = "m7b5"  // 半减七和弦
	ChordModeDim7    ChordMode = "dim7"  // 减七和弦
	ChordModePM7     ChordMode = "+M7"   // 增大七和弦
	ChordModeP7      ChordMode = "+7"    // 增七和弦
	ChordMode7sus4   ChordMode = "7sus4" // 属七挂四和弦
	ChordMode7b5     ChordMode = "7b5"   // 属七降五和弦
	ChordModeMinMaj7 ChordMode = "mM7"   // 小大七和弦
	ChordModeM7s5    ChordMode = "M7#5"  // 大七升五和弦
	ChordMode7s5     ChordMode = "7#5"   // 属七升五和弦
	ChordMode7b9     ChordMode = "7b9"   // 属七降九和弦
	ChordMode7s9     ChordMode = "7#9"   // 属七升九和弦

	// 九和弦

	ChordMode9       ChordMode = "9"     // 属九和弦
	ChordModeMaj9    ChordMode = "maj9"  // 大九和弦
	ChordModeMin9    ChordMode = "m9"    // 小九和弦
	ChordModeM9s5    ChordMode = "9#5"   // 属九升五和弦
	ChordModeMinMaj9 ChordMode = "mM9"   // 小大九和弦
	ChordMode9b5     ChordMode = "9b5"   // 属九降五和弦
	ChordMode9sus4   ChordMode = "9sus4" // 属九挂四和弦

	// 十一和弦

	ChordMode11    ChordMode = "11"    // 属十一和弦
	ChordModeMin11 ChordMode = "m11"   // 小十一和弦
	ChordModeMaj11 ChordMode = "maj11" // 大十一和弦
	ChordMode7s11  ChordMode = "7#11"  // 属七升十一和弦
	ChordMode9s11  ChordMode = "9#11"  // 属九升十一和弦
	ChordModeM7s11 ChordMode = "M7#11" // 大七升十一和弦

	// 十三和弦

	ChordMode13     ChordMode = "13"     // 属十三和弦
	ChordModeMaj13  ChordMode = "maj13"  // 大十三和弦
	ChordModeMin13  ChordMode = "m13"    // 小十三和弦
	ChordMode13sus4 ChordMode = "13sus4" // 属十三挂四和弦
	ChordMode7b13   ChordMode = "7b13"   // 属七降十三和弦

	// 兼容别名

	ChordModeMajor   = ChordModeMaj  // 大三和弦（ChordModeMaj 的别名）
	ChordModeMinor   = ChordModeMin  // 小三和弦（ChordModeMin 的别名）
	ChordModePower   = ChordMode5    // 强力和弦（ChordMode5 的别名）
	ChordModeMaj6    = ChordModeM6   // 大六和弦（ChordModeM6 的别名）
	ChordModeMin7b5  = ChordModeM7b5 // 半减七和弦（ChordModeM7b5 的别名）
	ChordModeAug7    = ChordModeP7   // 增七和弦（ChordModeP7 的别名）
	ChordModeAugMaj7 = ChordModePM7  // 增大七和弦（ChordModePM7 的别名）
	ChordModeMaj6p9  = ChordMode6p9  // 六九和弦（ChordMode6p9 的别名）
	ChordModeMin6p9  = ChordModeM6p9 // 小六九和弦（ChordModeM6p9 的别名）
)

type ChordTemplate struct {
	ToneSet

	Mode ChordMode
}

type ChordInstance struct {
	Template ChordTemplate // 模板
	Root     Tone          // 根音
}

//--------------------------------//
// ChordBuilder                   //
//--------------------------------//

type ChordBuilder struct {
	mode ChordMode
	root Tone
}

func (inst *ChordBuilder) BuildTemplate() *ChordTemplate {

	dst := new(ChordTemplate)

	mode := ChordModeMaj
	if inst != nil && inst.mode != "" {
		mode = inst.mode
	}

	dst.Mode = mode

	builder := new(ToneSetBuilder)
	intervals := chordIntervals(mode)
	dst.ToneSet = *builder.SetPattern(buildIntervalPattern(intervals)).Build(nil)

	return dst
}

func (inst *ChordBuilder) BuildInstance() *ChordInstance {

	dst := new(ChordInstance)

	dst.Template = *inst.BuildTemplate()

	var root Tone
	if inst != nil {
		root = inst.root
	}
	root.Name = root.Number.Name()
	root.Offset = 0
	dst.Root = root
	dst.Root.Base = &dst.Root

	return dst
}

// SetMode 设置要构建的和弦类型。
func (inst *ChordBuilder) SetMode(mode ChordMode) *ChordBuilder {
	if inst != nil {
		inst.mode = mode
	}
	return inst
}

// SetRoot 设置和弦实例的根音。
func (inst *ChordBuilder) SetRoot(root Tone) *ChordBuilder {
	if inst != nil {
		inst.root = root
	}
	return inst
}

// chordIntervals 返回各和弦相对根音的半音间隔。
func chordIntervals(mode ChordMode) []int {
	switch mode {
	case ChordModeMaj:
		return []int{0, 4, 7}
	case ChordModeMin:
		return []int{0, 3, 7}
	case ChordModeAug:
		return []int{0, 4, 8}
	case ChordModeDim:
		return []int{0, 3, 6}
	case ChordModeSus2:
		return []int{0, 2, 7}
	case ChordModeSus4:
		return []int{0, 5, 7}
	case ChordMode5:
		return []int{0, 7}
	case ChordModeM6:
		return []int{0, 4, 7, 9}
	case ChordModeMin6:
		return []int{0, 3, 7, 9}
	case ChordMode6p9:
		return []int{0, 4, 7, 9, 14}
	case ChordModeM6p9:
		return []int{0, 3, 7, 9, 14}
	case ChordModeAdd9:
		return []int{0, 4, 7, 14}
	case ChordModeMinAdd9:
		return []int{0, 3, 7, 14}
	case ChordModeMaj7:
		return []int{0, 4, 7, 11}
	case ChordModeMin7:
		return []int{0, 3, 7, 10}
	case ChordMode7:
		return []int{0, 4, 7, 10}
	case ChordModeM7b5:
		return []int{0, 3, 6, 10}
	case ChordModeDim7:
		return []int{0, 3, 6, 9}
	case ChordModePM7:
		return []int{0, 4, 8, 11}
	case ChordModeP7:
		return []int{0, 4, 8, 10}
	case ChordMode7sus4:
		return []int{0, 5, 7, 10}
	case ChordMode7b5:
		return []int{0, 4, 6, 10}
	case ChordModeMinMaj7:
		return []int{0, 3, 7, 11}
	case ChordModeM7s5:
		return []int{0, 4, 8, 11}
	case ChordMode7s5:
		return []int{0, 4, 8, 10}
	case ChordMode7b9:
		return []int{0, 4, 7, 10, 13}
	case ChordMode7s9:
		return []int{0, 4, 7, 10, 15}
	case ChordMode9:
		return []int{0, 4, 7, 10, 14}
	case ChordModeMaj9:
		return []int{0, 4, 7, 11, 14}
	case ChordModeMin9:
		return []int{0, 3, 7, 10, 14}
	case ChordModeM9s5:
		return []int{0, 4, 8, 10, 14}
	case ChordModeMinMaj9:
		return []int{0, 3, 7, 11, 14}
	case ChordMode9b5:
		return []int{0, 4, 6, 10, 14}
	case ChordMode9sus4:
		return []int{0, 5, 7, 10, 14}
	case ChordMode11:
		return []int{0, 4, 7, 10, 14, 17}
	case ChordModeMin11:
		return []int{0, 3, 7, 10, 14, 17}
	case ChordModeMaj11:
		return []int{0, 4, 7, 11, 14, 17}
	case ChordMode7s11:
		return []int{0, 4, 7, 10, 18}
	case ChordMode9s11:
		return []int{0, 4, 7, 10, 14, 18}
	case ChordModeM7s11:
		return []int{0, 4, 7, 11, 18}
	case ChordMode13:
		return []int{0, 4, 7, 10, 14, 21}
	case ChordModeMaj13:
		return []int{0, 4, 7, 11, 14, 21}
	case ChordModeMin13:
		return []int{0, 3, 7, 10, 14, 21}
	case ChordMode13sus4:
		return []int{0, 5, 7, 10, 14, 21}
	case ChordMode7b13:
		return []int{0, 4, 7, 10, 20}
	default:
		return nil
	}
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

	// 教会调式

	ScaleModeIonian     ScaleMode = "ionian"     // 伊奥尼亚调式
	ScaleModeDorian     ScaleMode = "dorian"     // 多利亚调式
	ScaleModePhrygian   ScaleMode = "phrygian"   // 弗里几亚调式
	ScaleModeLydian     ScaleMode = "lydian"     // 利底亚调式
	ScaleModeMixolydian ScaleMode = "mixolydian" // 混合利底亚调式
	ScaleModeAeolian    ScaleMode = "aeolian"    // 爱奥利亚调式
	ScaleModeLocrian    ScaleMode = "locrian"    // 洛克里亚调式

	// 五声音阶

	ScaleModePentatonic      ScaleMode = "Pentatonic"      // 五声音阶
	ScaleModeMajorPentatonic ScaleMode = "MajorPentatonic" // 大调五声音阶
	ScaleModeMinorPentatonic ScaleMode = "MinorPentatonic" // 小调五声音阶

	// 其它常用音阶

	ScaleModeBlues      ScaleMode = "blues"       // 小调布鲁斯音阶
	ScaleModeMajorBlues ScaleMode = "major_blues" // 大调布鲁斯音阶
	ScaleModeWholeTone  ScaleMode = "whole_tone"  // 全音阶
	ScaleModeChromatic  ScaleMode = "chromatic"   // 半音阶
)

type ScaleTemplate struct {
	ToneSet

	Mode ScaleMode
}

type ScaleInstance struct {
	Template ScaleTemplate // 模板
	Primary  Tone          // 基调 (主音)
}

//--------------------------------//
// ScaleBuilder                   //
//--------------------------------//

type ScaleBuilder struct {
}

func (inst *ScaleBuilder) BuildTemplate(mode ScaleMode) *ScaleTemplate {

	dst := new(ScaleTemplate)
	dst.Mode = mode

	builder := new(ToneSetBuilder)
	pattern := buildScalePattern(mode)
	dst.ToneSet = *builder.SetPattern(pattern).Build(nil)

	return dst
}

func (inst *ScaleBuilder) BuildInstance(mode ScaleMode, primary Tone) *ScaleInstance {

	dst := new(ScaleInstance)
	dst.Template = *inst.BuildTemplate(mode)
	dst.Primary = primary

	return dst
}

// scaleIntervals 返回各调式相对主音的半音间隔。
func scaleIntervals(mode ScaleMode) []int {
	switch mode {
	case ScaleModeMajor, ScaleModeNaturalMajor:
		return []int{0, 2, 4, 5, 7, 9, 11}
	case ScaleModeMinor, ScaleModeNaturalMinor:
		return []int{0, 2, 3, 5, 7, 8, 10}
	case ScaleModeHarmonicMinor:
		return []int{0, 2, 3, 5, 7, 8, 11}
	case ScaleModeMelodicMinor:
		return []int{0, 2, 3, 5, 7, 9, 11}
	case ScaleModeIonian:
		return []int{0, 2, 4, 5, 7, 9, 11}
	case ScaleModeDorian:
		return []int{0, 2, 3, 5, 7, 9, 10}
	case ScaleModePhrygian:
		return []int{0, 1, 3, 5, 7, 8, 10}
	case ScaleModeLydian:
		return []int{0, 2, 4, 6, 7, 9, 11}
	case ScaleModeMixolydian:
		return []int{0, 2, 4, 5, 7, 9, 10}
	case ScaleModeAeolian:
		return []int{0, 2, 3, 5, 7, 8, 10}
	case ScaleModeLocrian:
		return []int{0, 1, 3, 5, 6, 8, 10}
	case ScaleModePentatonic, ScaleModeMajorPentatonic:
		return []int{0, 2, 4, 7, 9}
	case ScaleModeMinorPentatonic:
		return []int{0, 3, 5, 7, 10}
	case ScaleModeBlues:
		return []int{0, 3, 5, 6, 7, 10}
	case ScaleModeMajorBlues:
		return []int{0, 2, 3, 4, 7, 9}
	case ScaleModeWholeTone:
		return []int{0, 2, 4, 6, 8, 10}
	case ScaleModeChromatic:
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	default:
		return nil
	}
}

// buildScalePattern 将半音间隔转换成 ToneSetBuilder 使用的点位模式。
func buildScalePattern(mode ScaleMode) string {
	return buildIntervalPattern(scaleIntervals(mode))
}

// buildIntervalPattern 将半音间隔转换成 ToneSetBuilder 使用的点位模式。
func buildIntervalPattern(intervals []int) string {
	if len(intervals) == 0 {
		return ""
	}

	pattern := make([]rune, intervals[len(intervals)-1]+1)
	for i := range pattern {
		pattern[i] = '.'
	}
	for i, interval := range intervals {
		pattern[interval] = rune('a' + i)
	}
	return string(pattern)
}

////////////////////////////////////////////////////////////////////////////////
// note

// MIDI 按键编码, 例如:(value=69,A4,440Hz), (value=60,C4(middle_C),261.6Hz)
type NotePitch = PitchNumber

////////////////////////////////////////////////////////////////////////////////
// EOF
