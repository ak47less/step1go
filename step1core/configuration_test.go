package step1core

import "testing"

func TestEngineConfigurationLS(t *testing.T) {

	ls := new(EngineConfigurationLS)
	cfg := new(EngineConfiguration)
	cfg2 := new(EngineConfiguration)

	elem1 := new(ElementConfig)
	wire1 := new(WireConfig)

	cfg.Audio.BufferLength = 1024
	cfg.Audio.SampleRate = 44100

	cfg.MIDI.BPM = 78.9
	cfg.MIDI.TPQN = 480
	cfg.MIDI.Tempo.Upper = 3
	cfg.MIDI.Tempo.Lower = 4

	cfg.Elements = append(cfg.Elements, elem1)

	elem1.ID = "mock1"
	elem1.Name = "Mock_1"
	elem1.Class = "class/mock"
	elem1.Wires = append(elem1.Wires, wire1)

	wire1.Destination.Element = "a"
	wire1.Destination.Port = "b"
	wire1.Source.Element = "c"
	wire1.Source.Port = "d"

	data, err := ls.EncodeJSON(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	err = ls.DecodeJSON(data, cfg2)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("engine.config.json = %s", string(data))

}
