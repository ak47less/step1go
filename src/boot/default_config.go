package boot

import "github.com/ak47less/step1go"

func GetDefaultConfig() *step1go.EngineConfiguration {

	cfg := new(step1go.EngineConfiguration)

	cfg.MIDI.BPM = 99
	cfg.MIDI.TPQN = 480
	cfg.Audio.SampleRate = 44100

	return cfg
}
