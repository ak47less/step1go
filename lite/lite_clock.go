package lite

import (
	"errors"
	"reflect"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type ClockEvent struct {
	Clock   Clock
	Time    AbsolutePosition
	Context *Context
}

type ClockListener interface {
	OnClock(ce *ClockEvent) error
}

////////////////////////////////////////////////////////////////////////////////

type Clock interface {
	AddListener(l ClockListener)

	RemoveListener(l ClockListener)

	Dispatch(ce *ClockEvent) error
}

////////////////////////////////////////////////////////////////////////////////

/**************************/
/*  class SimpleClock           */
/**************************/

type SimpleClock struct {
	mu        sync.RWMutex
	listeners []ClockListener
}

func (inst *SimpleClock) AddListener(l ClockListener) {
	if isNilClockListener(l) {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	for _, item := range inst.listeners {
		if sameClockListener(item, l) {
			return
		}
	}
	inst.listeners = append(inst.listeners, l)
}

func (inst *SimpleClock) RemoveListener(l ClockListener) {
	if isNilClockListener(l) {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	listeners := inst.listeners[:0]
	for _, item := range inst.listeners {
		if !sameClockListener(item, l) {
			listeners = append(listeners, item)
		}
	}
	inst.listeners = listeners
}

func (inst *SimpleClock) innerGetNow(ce *ClockEvent) AbsolutePosition {

	if ce == nil {
		return 0
	}

	ctx := ce.Context
	if ctx == nil {
		return 0
	}

	ac := ctx.AudioContext
	if ac == nil {
		return 0
	}

	sec := ac.GetCurrentTime()
	return AbsolutePosition(sec)
}

func (inst *SimpleClock) innerUpdateRTC(ce *ClockEvent) {

	var raw, normal, fine StepTickPos
	var root_pos, work_pos RelativePosition

	ctx := ce.Context
	rtc := &ctx.RTC
	work := ctx.WorkTileline
	root := ctx.RootTileline
	sc := ctx.SequencerContext
	info := &sc.Info
	sec_per_step := info.SecondsPerStep
	step_per_cyc := info.StepsPerCycle

	now := inst.innerGetNow(ce) // abs-time

	now.Relative(root, &root_pos)
	now.Relative(work, &work_pos)

	if sec_per_step < 0.001 {
		sec_per_step = 0.001
	}

	if step_per_cyc < 4 {
		step_per_cyc = 4
	}

	// steps

	var step_count int
	step_count = int(work_pos.Pos / sec_per_step)

	if step_count < 0 {
		step_count = 0
	}

	location := StepLocation(step_count)
	index := StepIndex(step_count % int(step_per_cyc))

	//raw

	raw.Index = index
	raw.Location = location
	raw.Position = root_pos

	//fine

	fine.Index = index
	fine.Location = location
	fine.Position = work_pos

	//normal

	normal.Index = index
	normal.Location = location
	normal.Position = work_pos

	normal.Position.Pos = sec_per_step * Seconds(step_count)

	// save

	ce.Time = now

	rtc.Raw = raw
	rtc.Precision = fine
	rtc.Normal = normal

}

func (inst *SimpleClock) Dispatch(ce *ClockEvent) error {
	if ce == nil {
		return errors.New("lite: nil clock event")
	}

	inst.mu.RLock()
	listeners := append([]ClockListener(nil), inst.listeners...)
	inst.mu.RUnlock()

	inst.innerUpdateRTC(ce)
	ce.Clock = inst

	var firstErr error
	var joinedErr error
	for _, listener := range listeners {
		err := listener.OnClock(ce)
		if err == nil {
			continue
		}
		if firstErr == nil {
			firstErr = err
		} else {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	if joinedErr != nil {
		return errors.Join(firstErr, joinedErr)
	}
	return firstErr
}

func isNilClockListener(l ClockListener) bool {
	if l == nil {
		return true
	}

	value := reflect.ValueOf(l)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func sameClockListener(a ClockListener, b ClockListener) bool {
	if isNilClockListener(a) || isNilClockListener(b) {
		return false
	}

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb || !ta.Comparable() {
		return false
	}
	return a == b
}

func (inst *SimpleClock) _impl() Clock {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
