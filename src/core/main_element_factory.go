package core

import "github.com/ak47less/step1go"

type MainElementFactory struct {
	table map[step1go.ElementClass]*step1go.ElementRegistration
}

// Register implements [step1go.ElementRegistry].
func (inst *MainElementFactory) Register(provider step1go.ElementProvider) {
	panic("unimplemented")
}

// CreateElement implements [step1go.ElementFactory].
func (inst *MainElementFactory) CreateElement(ei *step1go.ElementInfo) (step1go.Element, error) {
	panic("unimplemented")
}

func (inst *MainElementFactory) _impl() (step1go.ElementFactory, step1go.ElementRegistry) {
	return inst, inst
}
