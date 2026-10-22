package boot

import (
	"github.com/ak47less/step1go/step1core"
)

type Runner struct {
	Config *step1core.EngineConfiguration
	Loader *Loader
}

func (inst *Runner) Run() error {

	ldr := inst.Loader
	cfg := inst.Config

	if ldr == nil {
		ldr = inst.innerGetDefaultLoader()
	}

	if cfg == nil {
		cfg = inst.innerGetDefaultConfig()
	}

	en, err := ldr.Load(cfg)
	if err != nil {
		return err
	}

	err = en.Start()
	if err != nil {
		return err
	}

	err = en.Join()
	if err != nil {
		return err
	}

	return nil
}

func (inst *Runner) innerGetDefaultLoader() *Loader {
	return new(Loader)
}

func (inst *Runner) innerGetDefaultConfig() *step1core.EngineConfiguration {
	return GetDefaultConfig()
}
