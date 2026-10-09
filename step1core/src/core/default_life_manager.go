package core

import "github.com/ak47less/step1go/step1core"

type DefaultLifeManager struct {
}

// GetMain implements [step1core.LifeManager].
func (inst *DefaultLifeManager) GetMain() step1core.LifeAPI {
	panic("unimplemented")
}

// Add implements [step1core.LifeManager].
func (inst *DefaultLifeManager) Add(l step1core.LifeAPI) {
	panic("unimplemented")
}

// ListAll implements [step1core.LifeManager].
func (inst *DefaultLifeManager) ListAll() []*step1core.Life {
	panic("unimplemented")
}

func (inst *DefaultLifeManager) _impl() step1core.LifeManager {
	return inst
}
