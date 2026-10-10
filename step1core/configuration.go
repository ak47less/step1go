package step1core

import "encoding/json"

type AudioConfig struct {

	// samples per second
	SampleRate uint `json:"sample_rate"`

	// samples per buffer
	BufferLength uint `json:"buffer_length"`
}

type MidiConfig struct {

	// 每分钟的拍子数量
	BPM BPM `json:"bpm"`

	// ticks per quarter-note
	TPQN uint `json:"tpqn"`

	Tempo Tempo `json:"tempo"`
}

type ElementConfig struct {
	ID ElementID `json:"id"`

	Name ElementName `json:"name"`

	Class ElementClass `json:"class"`

	Wires []*WireConfig `json:"wires"`
}

type EngineConfiguration struct {
	Audio AudioConfig `json:"audio"`

	MIDI MidiConfig `json:"midi"`

	Elements []*ElementConfig `json:"elements"`
}

////////////////////////////////////////////////////////////////////////////////

type EngineConfigurationLS struct {
}

func (inst *EngineConfigurationLS) DecodeJSON(j []byte, cfg *EngineConfiguration) error {
	return json.Unmarshal(j, cfg)
}

func (inst *EngineConfigurationLS) EncodeJSON(cfg *EngineConfiguration) ([]byte, error) {
	const (
		prefix = ""
		indent = "\t"
	)
	return json.MarshalIndent(cfg, prefix, indent)
}

////////////////////////////////////////////////////////////////////////////////
