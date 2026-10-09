package boot

import "github.com/ak47less/step1go/step1core"

func GetDefaultConfig() *step1core.EngineConfiguration {

	cfg := new(step1core.EngineConfiguration)

	cfg.MIDI.BPM = 99
	cfg.MIDI.TPQN = 480
	cfg.Audio.SampleRate = 44100

	return cfg
}
