package common

////////////////////////////////////////////////////////////////////////////////

type Action int

const (
	ActionMin = iota

	ActionStartEngine
	ActionStopEngine

	ActionPlay
	ActionLoop
	ActionStop
	ActionPause
	ActionResume

	ActionNoteOn
	ActionNoteOff

	ActionMax
)

////////////////////////////////////////////////////////////////////////////////

type ActionEvent struct {
	Action Action

	FlagControl bool // as a Control action
	FlagNote    bool // as a MIDI-Note action
	FlagKeyUp   bool
	FlagKeyDown bool

	// for note-action

	NoteNumber   MidiNoteNumber
	NoteVelocity MidiNoteVelocity
}
