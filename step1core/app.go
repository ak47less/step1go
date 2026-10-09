package step1core

import "fmt"

type Application struct {
	config *EngineConfiguration

	engine Engine

	engineFactory EngineFactory
}

func (inst *Application) Run() error {

	engine, err := inst.NewEngine()
	if err != nil {
		return err
	}

	inst.engine = engine

	steps := make([]func() error, 0)
	steps = append(steps, engine.Create)
	steps = append(steps, engine.Start)
	steps = append(steps, engine.Run)
	steps = append(steps, engine.Stop)
	steps = append(steps, engine.Destroy)

	for _, fn := range steps {
		err = fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Application) NewEngine() (Engine, error) {

	cfg, err := inst.innerGetConfig()
	if err != nil {
		return nil, err
	}

	factory, err := inst.innerGetFactory()
	if err != nil {
		return nil, err
	}

	return factory.CreateEngine(cfg)
}

func (inst *Application) innerGetConfig() (*EngineConfiguration, error) {
	cfg := inst.config
	if cfg == nil {
		return nil, fmt.Errorf("EngineConfiguration is nil")
	}
	return cfg, nil
}

func (inst *Application) innerGetFactory() (EngineFactory, error) {
	f := inst.engineFactory
	if f == nil {
		return nil, fmt.Errorf("EngineFactory is nil")
	}
	return f, nil
}

func (inst *Application) SetConfiguration(cfg *EngineConfiguration) *Application {
	inst.config = cfg
	return inst
}

func (inst *Application) SetFactory(f EngineFactory) *Application {
	inst.engineFactory = f
	return inst
}
