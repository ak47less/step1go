package lite

import (
	"io"
	"os"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

type StdController struct {
	context  *EngineContext
	monitor  *Monitor
	handlers map[rune]*innerControllerHandlerInfo
}

func (inst *StdController) Start() {

	inst.init()

	go func() {
		err := inst.run()
		handleError(err)
	}()
}

func (inst *StdController) init() {

	inst.handlers = make(map[rune]*innerControllerHandlerInfo)

	inst.addHandler('=', "pause", inst.handleDoPause)
	inst.addHandler('p', "play", inst.handleDoPlay)
	inst.addHandler('r', "record", inst.handleDoRecord)
	inst.addHandler('-', "resume", inst.handleDoResume)
	inst.addHandler('s', "stop", inst.handleDoStop)

}

func (inst *StdController) addHandler(key rune, label string, fn innerControllerHandlerFunc) {
	item := &innerControllerHandlerInfo{
		key:     key,
		label:   label,
		handler: fn,
	}
	inst.handlers[key] = item
}

func (inst *StdController) run() error {

	var rdr io.Reader
	input := os.Stdin
	rdr = input
	buf := []byte{0}

	for {
		if inst.stopping() {
			break
		}
		cnt, err := rdr.Read(buf)
		if cnt == 1 && err == nil {
			k := buf[0]
			err = inst.handleInputKey(k)
			handleError(err)
		}
		time.Sleep(time.Millisecond * 10)
	}

	return nil
}

func (inst *StdController) stopping() bool {
	state := &inst.context.State
	return (state.Stopping || state.Stopped)
}

func (inst *StdController) handleInputKey(key byte) error {

	if key == 10 {
		return nil // skip 'enter'
	}

	mo := inst.monitor
	table := inst.handlers
	code := rune(key)
	item := table[code]
	ec := inst.context

	mo.InputKeyCode = code
	mo.InputKeyCount++

	if item != nil {
		return item.handler(ec)
	}

	return nil
}

func (inst *StdController) handleDoPlay(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Play()
}

func (inst *StdController) handleDoStop(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Stop()
}

func (inst *StdController) handleDoPause(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Pause()
}

func (inst *StdController) handleDoResume(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Resume()
}

func (inst *StdController) handleDoRecord(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Record()
}

////////////////////////////////////////////////////////////////////////////////
// EOF
