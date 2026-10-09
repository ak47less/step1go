package step1core

// MIDI 按键编码, 例如:(value=69,A4,440Hz), (value=60,C4(middle_C),261.6Hz)
type NoteNumber byte

type NoteName string

type NoteGroup int

type NoteInfo struct {
	Number NoteNumber

	Name NoteName

	Group NoteGroup

	Frequency Frequency
}
