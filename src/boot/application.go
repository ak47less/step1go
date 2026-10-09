package boot

import (
	"github.com/ak47less/step1go"
	"github.com/ak47less/step1go/src/core"
)

func NewApp() (*step1go.Application, error) {

	app := new(step1go.Application)
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
