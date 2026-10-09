package core

import "github.com/ak47less/step1go/step1core"

type DefaultEngineFactory struct {
}

// CreateEngine implements [step1core.EngineFactory].
func (inst *DefaultEngineFactory) CreateEngine(cfg *step1core.EngineConfiguration) (step1core.Engine, error) {

	var err error
	logger := step1core.Log()
	engine := new(innerDefaultEngine)
	ec := new(step1core.EngineContext)
	track_man := new(DefaultTrackManager)
	ele_man := new(DefaultElementManager)
	main_ef := new(MainElementFactory)
	loader := new(DefaultEngineLoader)
	life_man := new(DefaultLifeManager)

	ele_man.SetFactory(main_ef)

	ec.Engine = engine
	ec.Configuration = cfg
	ec.ElementFactory = main_ef
	ec.ElementManager = ele_man
	ec.ElementRegistry = main_ef
	ec.TrackManager = track_man
	ec.LifeManager = life_man
	ec.Loader = loader
	ec.Logger = logger

	engine.context = ec

	err = engine.Init(cfg)
	if err != nil {
		return nil, err
	}
	return engine, nil
}

func (inst *DefaultEngineFactory) _impl() step1core.EngineFactory {
	return inst
}
