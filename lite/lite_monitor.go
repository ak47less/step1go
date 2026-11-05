package lite

import (
	"fmt"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////

type monitoring struct {
	timer   *Timer
	context *Context
	logger  Logger
}

////////////////////////////////////////////////////////////////////////////////

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
	tasklist = append(tasklist, inst.innerLogSTC)
	tasklist = append(tasklist, inst.innerLogCube)

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

func (inst *Monitor) innerLogSTC(mo *monitoring) error {

	logger := mo.logger
	sc := mo.context.SequencerContext
	stc := sc.STC
	txt := ""

	if stc != nil {
		current := stc.Current
		pos := current.Position.Position
		loc := current.Location
		idx := current.Index
		txt = fmt.Sprintf("[STC(StepTickContext) location:%d index:%d pos:%f ]", loc, idx, pos)
	} else {
		txt = "STC is nil"
	}

	logger.Info(txt)
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

func (inst *Monitor) innerLogCube(mo *monitoring) error {

	var cml innerCubeMonitorLogger
	cml.log(mo)
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

////////////////////////////////////////////////////////////////////////////////
// innerCubeMonitorLogger

type innerCubeMonitorLogger struct {
	charCurrentSound rune
	charCurrentMute  rune
	charNormalSound  rune
	charNormalMute   rune
}

func (inst *innerCubeMonitorLogger) log(mo *monitoring) {

	inst.charCurrentMute = 'x'
	inst.charCurrentSound = '@'
	inst.charNormalMute = '_'
	inst.charNormalSound = 'o'

	sc := mo.context.SequencerContext
	bank, err := sc.CurrentBankHolder.GetBank()

	if err == nil && bank != nil {
		inst.logbank(mo, bank)
	}

}

func (inst *innerCubeMonitorLogger) logbank(mo *monitoring, bank *Bank) {

	context := mo.context
	sc := context.SequencerContext
	cube := &sc.Cube

	istep_now := sc.STC.Realtime.Index

	tracks := bank.Tracks
	builder := new(strings.Builder)
	step_cnt := cube.Steps.Length()
	track_cnt := cube.Tracks.Length()

	for _, track_item := range tracks {

		cells := &track_item.Cells
		builder.Reset()

		for _, step_cell := range cells.list {

			istep := step_cell.StepIndex
			if uint(istep) > step_cnt {
				break
			}

			ch := inst.innerGetCharForCell(step_cell, istep_now)
			builder.WriteRune(ch)
		}

		itrack := track_item.Index
		if uint(itrack) > track_cnt {
			break
		}

		builder.WriteString(fmt.Sprintf(" : track_%d", itrack))
		fmt.Println(builder.String())
	}

}

func (inst *innerCubeMonitorLogger) innerGetCharForCell(cell *SimpleCell, now StepIndex) rune {

	sound := false
	hit := (now == cell.StepIndex)

	if cell != nil {
		if cell.Enabled && (cell.Note != nil) {
			sound = true
		}
	}

	if sound {
		if hit {
			return inst.charCurrentSound
		} else {
			return inst.charNormalSound
		}
	} else {
		if hit {
			return inst.charCurrentMute
		} else {
			return inst.charNormalMute
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// EOF
