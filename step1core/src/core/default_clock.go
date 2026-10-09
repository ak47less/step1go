package core

import "github.com/ak47less/step1go/step1core"

type DefaultClock struct {
}

// Next implements [step1core.Clock].
func (inst *DefaultClock) Next(c *step1core.Bus) {
	panic("unimplemented")
}

func (inst *DefaultClock) _impl() step1core.Clock {
	return inst
}
