package app

import (
	"github.com/ak47less/step1go/lite2/core"
	"github.com/ak47less/step1go/lite2/etc/step1lite2"
	"github.com/ak47less/step1go/lite2/lib"
)

////////////////////////////////////////////////////////////////////////////////

type Configuration = lib.Configuration

////////////////////////////////////////////////////////////////////////////////

type Step1Lite2App struct {
	config *Configuration
}

func (inst *Step1Lite2App) Configure() *Configuration {
	return inst.innerGetConfig()
}

func (inst *Step1Lite2App) innerGetConfig() *Configuration {
	cfg := inst.config
	if cfg == nil {
		cfg = inst.innerLoadConfig()
		inst.config = cfg
	}
	return cfg
}

func (inst *Step1Lite2App) innerInitConfig(cfg *Configuration) {

	cfg.ApplicationContextFactory = lib.NewApplicationContextFactory()
	cfg.Configurators = make([]*lib.ConfiguratorRegistration, 0)
	cfg.Parts = make([]any, 0)
	cfg.Properties = make(core.Properties)

}

func (inst *Step1Lite2App) innerLoadConfig() *Configuration {

	cfg := new(Configuration)
	inst.innerInitConfig(cfg)

	step1lite2.UseDefaultConfig(cfg)
	step1lite2.UseFinalConfig(cfg)
	step1lite2.UseCliConfig(cfg)

	return cfg
}

func (inst *Step1Lite2App) Run() error {
	cfg := inst.Configure()
	factory := cfg.ApplicationContextFactory
	ac, err := factory.CreateApplicationContext(cfg)
	if err != nil {
		return err
	}
	runner := ac.Runner
	return runner.Run(ac)
}
