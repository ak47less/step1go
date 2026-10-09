package core

import "github.com/ak47less/step1go/step1core"

type MainElementFactory struct {
	table map[step1core.ElementClass]*step1core.ElementRegistration
}

// Register implements [step1core.ElementRegistry].
func (inst *MainElementFactory) Register(provider step1core.ElementProvider) {
	panic("unimplemented")
}

// CreateElement implements [step1core.ElementFactory].
func (inst *MainElementFactory) CreateElement(ei *step1core.ElementInfo) (step1core.Element, error) {
	panic("unimplemented")
}

func (inst *MainElementFactory) _impl() (step1core.ElementFactory, step1core.ElementRegistry) {
	return inst, inst
}
