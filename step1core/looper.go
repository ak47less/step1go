package step1core

import (
	"time"
)

type Looper struct {
	context *EngineContext

	id int64
}

func (inst *Looper) Equals(other *Looper) bool {
	if inst == nil || other == nil {
		return false
	}
	return (inst.id == other.id)
}

func (inst *Looper) Start() error {

	go inst.run()

	return nil
}

func (inst *Looper) Init(ctx *EngineContext) {
	t := time.Now()
	n := t.UnixMilli()
	inst.id = n
	inst.context = ctx
}

func (inst *Looper) run() {

	err := inst.runWithErr()

	if err != nil {
		Log().Error("%s", err.Error())
	}

}

func (inst *Looper) runWithErr() error {

	var err error
	var index int64
	ctx := inst.context
	bus := new(Bus)

	clock := ctx.Clock
	elist := ctx.ElementManager.ListAll()
	state := &ctx.State
	logger := ctx.Logger
	lp := new(Looping)

	bus.Context = ctx
	state.Started = true

	lp.logger = logger
	lp.state = state
	lp.context = ctx
	lp.elementList = elist
	lp.bus = bus
	lp.clock = clock

	defer func() {
		state.Stopped = true
	}()

	for index = 0; ; index++ {

		if state.Stopping {
			break
		}

		//update
		if lp.needUpdate() {
			err = lp.update()
			lp.handleError(err)
		}

		// clock
		err = lp.handleClock()
		lp.handleError(err)

		// I/O
		err = lp.handleIO()
		lp.handleError(err)

		// log
		lp.log(index)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Looping struct {
	elementList []Element

	wireList []*Wire

	logger *Logger

	context *EngineContext

	state *EngineState

	bus *Bus

	clock Clock
}

func (inst *Looping) log(index int64) error {

	logger.Trace("engine.loop[%d]", index)

}

func (inst *Looping) needUpdate() bool {

	redo := ctx.ConfigRedo

}

func (inst *Looping) update() error {

	if redo != nil {
		ctx.ConfigRedo = nil
		for _, el := range elist {
			err = el.Configure(ctx, redo)
			if err != nil {
				return err
			}
		}
	}

}

func (inst *Looping) handleClock() error {

	var err error
	all := inst.elementList
	bus := inst.bus
	cl := inst.clock

	cl.Next(bus)

	for _, el := range all {
		err = el.Clock(bus)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Looping) handleIO() error {

	all := inst.wireList

	for _, el := range all {
		err = el.IO(bus)
		if err != nil {
			return err
		}
	}

}

func (inst *Looping) handleError(err error) {
	if err == nil {
		return
	}
	inst.logger.Error("%s", err.Error())
}

////////////////////////////////////////////////////////////////////////////////
// EOF
