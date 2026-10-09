package core

import "github.com/ak47less/step1go"

type DefaultElementManager struct {
	factory step1go.ElementFactory
}

// Count implements [step1go.ElementManager].
func (inst *DefaultElementManager) Count() int {
	panic("unimplemented")
}

// FindElement implements [step1go.ElementManager].
func (inst *DefaultElementManager) FindElement(ei *step1go.ElementInfo) (step1go.Element, error) {
	panic("unimplemented")
}

// GetElement implements [step1go.ElementManager].
func (inst *DefaultElementManager) GetElement(index int) step1go.Element {
	panic("unimplemented")
}

// ListAll implements [step1go.ElementManager].
func (inst *DefaultElementManager) ListAll() []step1go.Element {
	panic("unimplemented")
}

// LoadElement implements [step1go.ElementManager].
func (inst *DefaultElementManager) LoadElement(ei *step1go.ElementInfo) (step1go.Element, error) {
	panic("unimplemented")
}

func (inst *DefaultElementManager) _impl() step1go.ElementManager {
	return inst
}
