package step1core

type Channel struct {
	Index uint

	Name string

	Type string

	// buffer:

	Audio *AudioBuffer

	MIDI *MidiBuffer
}
