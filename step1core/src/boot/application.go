package boot

import (
	"github.com/ak47less/step1go/step1core"
	"github.com/ak47less/step1go/step1core/src/core"
)

func NewApp() (*step1core.Application, error) {

	app := new(step1core.Application)
	cfg := GetDefaultConfig()
	factory := new(core.DefaultEngineFactory)

	app.SetConfiguration(cfg)
	app.SetFactory(factory)

	return app, nil
}

func Run() error {
	app, err := NewApp()
	if err != nil {
		return err
	}
	return app.Run()
}
