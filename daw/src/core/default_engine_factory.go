package core

import "github.com/ak47less/step1go/step1core"

type DefaultEngineFactory struct {
}

// CreateEngine implements [step1core.EngineFactory].
func (inst *DefaultEngineFactory) CreateEngine(cfg *step1core.EngineConfiguration) (step1core.Engine, error) {

	var err error

	logger := step1core.Log()
	engine := new(innerDefaultEngineFacade)
	ctx := new(step1core.EngineContext)
	track_man := new(DefaultTrackManager)
	ele_man := new(DefaultDeviceManager)
	// main_ef := new(MainDeviceFactory)
	ls := new(DefaultEngineLS)
	life_man := new(DefaultLifeManager)
	clock := new(DefaultClock)

	ele_man.Init(ctx)

	ctx.Configuration = cfg
	ctx.DeviceManager = ele_man
	ctx.DeviceFactory = cfg.DeviceProviders
	ctx.DeviceRegistry = cfg.DeviceProviders
	ctx.TrackManager = track_man
	ctx.LifeManager = life_man
	ctx.LS = ls
	ctx.Logger = logger

	ctx.Transport = nil
	ctx.Engine = engine
	ctx.Clock = clock

	engine.context = ctx

	err = engine.Init(cfg)
	if err != nil {
		return nil, err
	}
	return engine, nil
}

func (inst *DefaultEngineFactory) _impl() step1core.EngineFactory {
	return inst
}
