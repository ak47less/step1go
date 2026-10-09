package core

import "github.com/ak47less/step1go"

type DefaultLifeManager struct {
}

// GetMain implements [step1go.LifeManager].
func (inst *DefaultLifeManager) GetMain() step1go.LifeAPI {
	panic("unimplemented")
}

// Add implements [step1go.LifeManager].
func (inst *DefaultLifeManager) Add(l step1go.LifeAPI) {
	panic("unimplemented")
}

// ListAll implements [step1go.LifeManager].
func (inst *DefaultLifeManager) ListAll() []*step1go.Life {
	panic("unimplemented")
}

func (inst *DefaultLifeManager) _impl() step1go.LifeManager {
	return inst
}
