package core

import "github.com/ak47less/step1go/step1core"

type DefaultElementManager struct {
	factory step1core.ElementFactory
}

// Count implements [step1core.ElementManager].
func (inst *DefaultElementManager) Count() int {
	panic("unimplemented")
}

// FindElement implements [step1core.ElementManager].
func (inst *DefaultElementManager) FindElement(ei *step1core.ElementInfo) (step1core.Element, error) {
	panic("unimplemented")
}

// GetElement implements [step1core.ElementManager].
func (inst *DefaultElementManager) GetElement(index int) step1core.Element {
	panic("unimplemented")
}

// ListAll implements [step1core.ElementManager].
func (inst *DefaultElementManager) ListAll() []step1core.Element {
	panic("unimplemented")
}

// LoadElement implements [step1core.ElementManager].
func (inst *DefaultElementManager) LoadElement(ei *step1core.ElementInfo) (step1core.Element, error) {
	panic("unimplemented")
}

func (inst *DefaultElementManager) _impl() step1core.ElementManager {
	return inst
}
