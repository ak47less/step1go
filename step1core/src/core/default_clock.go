package core

import (
	"time"

	"github.com/ak47less/step1go/step1core"
)

type DefaultClock struct {
	count int
}

// Next implements [step1core.Clock].
func (inst *DefaultClock) Next(c *step1core.Bus) {
	// panic("unimplemented")

	inst.count++
	time.Sleep(time.Second)

	// logger := step1core.Log()
	// logger.Trace("clock.count = %d", inst.count)

}

func (inst *DefaultClock) _impl() step1core.Clock {
	return inst
}
