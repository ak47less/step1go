package step1core

////////////////////////////////////////////////////////////////////////////////

// beats per minute
type BPM float32

// 表示频率
type Frequency float32

type Tempo struct {
	Upper uint
	Lower uint
}

////////////////////////////////////////////////////////////////////////////////

// 表示音高
type Pitch int

// 音程
type PitchInterval int

////////////////////////////////////////////////////////////////////////////////

// 音调, [C(0), C#(1) , D(2) , ... , A(9) , A#(10) , B(11) ]
type Tone byte

type ToneName string

const (
	ToneNameC  ToneName = "C"
	ToneNameCS ToneName = "C#"

	ToneNameD  ToneName = "D"
	ToneNameDS ToneName = "D#"

	ToneNameE ToneName = "E"
	// ToneName ToneName = "E#"

	ToneNameF  ToneName = "F"
	ToneNameFS ToneName = "F#"

	ToneNameG  ToneName = "G"
	ToneNameGS ToneName = "G#"

	ToneNameA  ToneName = "A"
	ToneNameAS ToneName = "A#"

	ToneNameB ToneName = "B"
	// ToneName ToneName = "B"

)

type ToneMap struct {
	tones []Tone
}

////////////////////////////////////////////////////////////////////////////////

// 和弦
type ChordMap struct {
	ToneMap
}

// 调式
type ScaleMap struct {
	ToneMap
}

////////////////////////////////////////////////////////////////////////////////
