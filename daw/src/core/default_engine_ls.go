package core

import "github.com/ak47less/step1go/step1core"

type DefaultEngineLS struct {
}

// Load implements [step1core.EngineLS].
func (inst *DefaultEngineLS) Load(ctx *step1core.EngineContext) error {
	panic("unimplemented")
}

// Save implements [step1core.EngineLS].
func (inst *DefaultEngineLS) Save(ctx *step1core.EngineContext) error {
	panic("unimplemented")
}

func (inst *DefaultEngineLS) _impl() step1core.EngineLS {
	return inst
}
