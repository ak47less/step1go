package core

import "github.com/ak47less/step1go"

type DefaultEngineLoader struct {
}

// Load implements [step1go.EngineLoader].
func (inst *DefaultEngineLoader) Load(en step1go.Engine) error {
	panic("unimplemented")
}

func (inst *DefaultEngineLoader) _impl() step1go.EngineLoader {
	return inst
}
