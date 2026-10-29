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
	ActionRecord

	ActionNoteOn
	ActionNoteOff

	ActionMasterVolumeUp
	ActionMasterVolumeDown
	ActionMasterVolumeMute

	ActionKeyUp
	ActionKeyDown

	ActionMax
)

////////////////////////////////////////////////////////////////////////////////

type ActionEvent struct {

	// common

	Action Action

	// flags

	FlagControl bool // as a Control action
	FlagNote    bool // as a MIDI-Note action
	FlagNoteOn  bool // as a MIDI-Note-on action
	FlagNoteOff bool // as a MIDI-Note-off action
	FlagCKey    bool // as a computer-key action
	FlagKeyUp   bool // as a computer-key-up action
	FlagKeyDown bool // as a computer-key-down action

	// for computer-key-action

	KeyCode ComputerKeyCode

	// for note-action

	NoteNumber   MidiNoteNumber
	NoteVelocity MidiNoteVelocity
}
