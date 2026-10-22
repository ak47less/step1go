package lite

type Boot struct {
}

func (inst *Boot) Run(cfg *Configuration) error {

	ef := DefaultEngineFactory()

	if cfg == nil {
		cfg = new(Configuration)
		inst.loadDefaultConfig(cfg)
	}

	engine, err := ef.CreateEngine(cfg)
	if err != nil {
		return err
	}

	err = engine.Create()
	if err != nil {
		return err
	}

	err = engine.Start()
	if err != nil {
		return err
	}

	err = engine.Stop()
	if err != nil {
		return err
	}

	err = engine.Destroy()
	if err != nil {
		return err
	}

	return nil
}

func (inst *Boot) loadDefaultConfig(cfg *Configuration) {

	cfg.Driver = NewMockDriver()
	cfg.BPM = 99
	cfg.SampleRate = 44100

}

func Run(cfg *Configuration) error {
	b := new(Boot)
	return b.Run(cfg)
}
