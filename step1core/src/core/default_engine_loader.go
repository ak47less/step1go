package core

import "github.com/ak47less/step1go/step1core"

type DefaultEngineLoader struct {
}

// Load implements [step1core.EngineLoader].
func (inst *DefaultEngineLoader) Load(en step1core.Engine) error {
	panic("unimplemented")
}

func (inst *DefaultEngineLoader) _impl() step1core.EngineLoader {
	return inst
}
