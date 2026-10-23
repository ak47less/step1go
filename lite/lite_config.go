package lite

type Configuration struct {
	Driver Driver

	SampleRate int

	BPM BPM

	CountBankMax  uint
	CountStepMax  uint
	CountTrackMax uint
}

func DefaultConfig() *Configuration {

	cfg := new(Configuration)
	driver := NewMockDriver()

	cfg.SampleRate = 44100
	cfg.BPM = 100
	cfg.Driver = driver

	cfg.CountBankMax = 16
	cfg.CountStepMax = 32
	cfg.CountTrackMax = 16

	return cfg
}
