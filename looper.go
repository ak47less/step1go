package step1go

import "time"

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

	bus.Context = ctx

	state.Started = true

	defer func() {

		state.Stopped = true

	}()

	for index = 0; ; index++ {

		redo := ctx.ConfigRedo

		if state.Stopping {
			break
		}

		if redo != nil {
			ctx.ConfigRedo = nil
			for _, el := range elist {
				err = el.Configure(ctx, redo)
				if err != nil {
					return err
				}
			}
		}

		// clock

		clock.Next(bus)

		for _, el := range elist {
			err = el.Clock(bus)
			if err != nil {
				return err
			}
		}

		// I/O

		for _, el := range elist {
			err = el.IO(bus)
			if err != nil {
				return err
			}
		}

		logger.Trace("engine.loop[%d]", index)
	}

	return nil
}
