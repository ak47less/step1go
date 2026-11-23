package lib

import "github.com/ak47less/step1go/lite2/core"

type Configuration struct {
	Properties core.Properties

	UseRestApi bool

	Configurators []*ConfiguratorRegistration

	Parts []any

	ApplicationContextFactory ApplicationContextFactory

	Driver Driver
}

type Configurator interface {
	Configure(cfg *Configuration)
}

type ConfiguratorRegistration struct {
	Name string

	Description string

	Order int

	Configurator Configurator
}
