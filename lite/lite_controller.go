package lite

import (
	"io"
	"os"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

type innerControllerHandlerFunc func(ec *EngineContext) error

type innerControllerHandlerInfo struct {
	key     rune
	label   string
	handler innerControllerHandlerFunc
}

////////////////////////////////////////////////////////////////////////////////

type Controller struct {
	context  *EngineContext
	monitor  *Monitor
	handlers map[rune]*innerControllerHandlerInfo
}

func (inst *Controller) Start() {

	inst.init()

	go func() {
		err := inst.run()
		handleError(err)
	}()
}

func (inst *Controller) init() {

	inst.handlers = make(map[rune]*innerControllerHandlerInfo)

	inst.addHandler('=', "pause", inst.handleDoPause)
	inst.addHandler('p', "play", inst.handleDoPlay)
	inst.addHandler('r', "record", inst.handleDoRecord)
	inst.addHandler('-', "resume", inst.handleDoResume)
	inst.addHandler('s', "stop", inst.handleDoStop)

}

func (inst *Controller) addHandler(key rune, label string, fn innerControllerHandlerFunc) {
	item := &innerControllerHandlerInfo{
		key:     key,
		label:   label,
		handler: fn,
	}
	inst.handlers[key] = item
}

func (inst *Controller) run() error {

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

func (inst *Controller) stopping() bool {
	state := &inst.context.State
	return (state.Stopping || state.Stopped)
}

func (inst *Controller) handleInputKey(key byte) error {

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

func (inst *Controller) handleDoPlay(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Play()
}

func (inst *Controller) handleDoStop(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Stop()
}

func (inst *Controller) handleDoPause(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Pause()
}

func (inst *Controller) handleDoResume(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Resume()
}

func (inst *Controller) handleDoRecord(ec *EngineContext) error {
	tr := ec.Transport
	return tr.Record()
}

////////////////////////////////////////////////////////////////////////////////
// EOF
