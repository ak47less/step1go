package step1core

type AudioConfig struct {

	// samples per second
	SampleRate uint

	// samples per buffer
	BufferLength uint
}

type MidiConfig struct {

	// 每分钟的拍子数量
	BPM BPM

	// ticks per quarter-note
	TPQN uint

	Tempo Tempo
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
