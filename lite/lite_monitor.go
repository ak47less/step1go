package lite

import "fmt"

type monitoring struct {
	timer   *Timer
	context *Context
	logger  Logger
}

type Monitor struct {
	context       *EngineContext
	fps           int
	InputKeyCount int
	InputKeyCode  rune
}

func (inst *Monitor) Init(ec *EngineContext) {
	inst.context = ec
	inst.fps = 24
}

func (inst *Monitor) Start() {

	tm := &Timer{
		Delay:    1.0,
		Interval: 1.0 / Seconds(inst.fps),
		Handler:  inst.ontimer,
	}

	mocfg := inst.context.Configuration.Monitor
	if mocfg == nil {
		return
	}
	mocfg = mocfg.Prepare()

	var ctrl Controller

	if mocfg.Scheme == "tcp" {
		ctrl = &TcpController{
			context: inst.context,
			monitor: inst,
		}
	} else {
		ctrl = &StdController{
			context: inst.context,
			monitor: inst,
		}
	}

	ctrl.Start()
	tm.Start()
}

func (inst *Monitor) ontimer(tm *Timer) error {

	tasklist := make([]func(*monitoring) error, 0)
	ec := inst.context

	mo := &monitoring{
		context: ec,
		logger:  ec.Logger,
		timer:   tm,
	}

	tasklist = append(tasklist, inst.innerLogBar)
	tasklist = append(tasklist, inst.innerLogEngineState)
	tasklist = append(tasklist, inst.innerLogTransportState)
	tasklist = append(tasklist, inst.innerLogPosition)
	tasklist = append(tasklist, inst.innerLogKeyInput)

	for _, task := range tasklist {
		err := task(mo)
		if err != nil {
			mo.logger.Error(err.Error())
		}
	}

	return nil
}

func (inst *Monitor) innerLogEngineState(mo *monitoring) error {

	ec := mo.context
	state := ec.State
	logger := mo.logger

	logger.Info("%s", state.String())

	if state.Stopped || state.Stopping {
		mo.timer.Stop()
	}

	return nil
}

func (inst *Monitor) innerLogTransportState(mo *monitoring) error {

	ec := inst.context
	tc := ec.TransportContext
	logger := mo.logger

	pa := tc.paused
	pl := tc.playing
	re := tc.recording

	logger.Info("transport.[state playing:%v recording:%v paused:%v]", pl, re, pa)

	return nil
}

func (inst *Monitor) innerLogPosition(mo *monitoring) error {

	ec := mo.context
	logger := mo.logger

	raw := ec.AudioContext.GetCurrentTime()
	root := ec.CurrentAtRoot
	work := ec.CurrentAtWork

	if root == nil || work == nil {
		return fmt.Errorf("context.CurrentAtRoot|context.CurrentAtWork is nil")
	}

	logger.Info("transport.[position raw:%f root:%f work:%f]", raw, root.Position, work.Position)

	return nil
}

func (inst *Monitor) innerLogKeyInput(mo *monitoring) error {

	const f = "input.[key count:%d code:%d char:'%s']"

	logger := mo.logger
	code := inst.InputKeyCode
	count := inst.InputKeyCount

	logger.Info(f, count, int(code), string(code))
	return nil
}

func (inst *Monitor) innerLogBar(mo *monitoring) error {
	logger := mo.logger
	const bar = "--------------------------------------------------------------"
	logger.Info(bar)
	return nil
}
