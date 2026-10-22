package lite

func DefaultEngineFactory() EngineFactory {
	return new(innerEngineFactory)
}

type innerEngineFactory struct {
}

// CreateEngine implements [EngineFactory].
func (inst *innerEngineFactory) CreateEngine(cfg *Configuration) (Engine, error) {
	panic("unimplemented")
}

func (inst *innerEngineFactory) _impl() EngineFactory {
	return inst
}
