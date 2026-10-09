package step1go

type AudioConfig struct {

	// samples per second
	SampleRate uint

	// samples per buffer
	BufferLength uint
}

type MidiConfig struct {

	// beats per minute
	BPM float32

	//  ticks per quarter-note
	TPQN uint
}

type ElementConfig struct {
	ID ElementID

	Name ElementName

	Class ElementClass

	Wires []*WireConfig
}

type EngineConfiguration struct {
	Audio AudioConfig

	MIDI MidiConfig

	Elements []*ElementConfig
}
