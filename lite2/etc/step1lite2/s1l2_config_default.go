package step1lite2

import (
	"github.com/ak47less/step1go/lite2/lib"
	"github.com/ak47less/step1go/lite2/webaudioapi/mock"
)

func UseDefaultConfig(cfg *lib.Configuration) {

	cr := &lib.ConfiguratorRegistration{
		Name:        "default_config",
		Description: "",
		Order:       0,
	}

	cr.Configurator = new(innerDefaultConfig)
	cfg.Configurators = append(cfg.Configurators, cr)
}

////////////////////////////////////////////////////////////////////////////////

type innerDefaultConfig struct {
}

// Configure implements [lib.Configurator].
func (i *innerDefaultConfig) Configure(cfg *lib.Configuration) {

	// driver

	cfg.Driver = mock.NewMockDriver()

	// props

	cfg.Properties.SetValue("config.default.done", "yes")

	// parts

	cfg.Parts = append(cfg.Parts, new(lib.DefaultAppRunner))
	cfg.Parts = append(cfg.Parts, new(lib.DefaultEngine))
	cfg.Parts = append(cfg.Parts, new(lib.DefaultEventDispatcher))
	cfg.Parts = append(cfg.Parts, new(lib.DefaultSequencer))
	cfg.Parts = append(cfg.Parts, new(lib.DefaultSession))

}

////////////////////////////////////////////////////////////////////////////////
