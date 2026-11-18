package step1lite2

import "github.com/ak47less/step1go/lite2/lib"

func UseFinalConfig(cfg *lib.Configuration) {

	cr := &lib.ConfiguratorRegistration{
		Name:        "final_config",
		Description: "",
		Order:       0,
	}

	cr.Configurator = new(innerFinalConfig)
	cfg.Configurators = append(cfg.Configurators, cr)
}

////////////////////////////////////////////////////////////////////////////////

type innerFinalConfig struct {
}

// Configure implements [lib.Configurator].
func (i *innerFinalConfig) Configure(cfg *lib.Configuration) {

	cfg.Properties.SetValue("config.final.done", "yes")

}

////////////////////////////////////////////////////////////////////////////////
