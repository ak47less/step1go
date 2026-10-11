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

type DeviceConfig struct {
	ID DeviceID `json:"id"`

	Name DeviceName `json:"name"`

	Class DeviceClass `json:"class"`

	Wires []*WireConfig `json:"wires"`
}

type EngineConfiguration struct {
	Audio AudioConfig `json:"audio"`

	MIDI MidiConfig `json:"midi"`

	Devices []*DeviceConfig `json:"devices"`
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
