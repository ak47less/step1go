package lite

type Configuration struct {
	Driver Driver

	SampleRate int

	BPM   BPM
	Tempo Tempo

	CountBankMax  uint
	CountStepMax  uint
	CountTrackMax uint

	Monitor bool
}

func DefaultConfig() *Configuration {

	cfg := new(Configuration)
	driver := NewMockDriver()

	cfg.Driver = driver
	cfg.SampleRate = 44100
	cfg.BPM = 100
	cfg.Tempo.Upper = 4
	cfg.Tempo.Lower = 4

	cfg.CountBankMax = 16
	cfg.CountStepMax = 32
	cfg.CountTrackMax = 16

	cfg.Monitor = false

	return cfg
}
