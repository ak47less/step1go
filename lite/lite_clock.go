package lite

import (
	"errors"
	"reflect"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type ClockEvent struct {
	Clock   *Clock
	Time    *RelativePosition
	Context *Context
}

type ClockListener interface {
	OnClock(ce *ClockEvent) error
}

/**************************/
/*  class Clock           */
/**************************/

type Clock struct {
	mu        sync.RWMutex
	listeners []ClockListener
}

func (inst *Clock) AddListener(l ClockListener) {
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

func (inst *Clock) RemoveListener(l ClockListener) {
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

func (inst *Clock) Dispatch(ce *ClockEvent) error {
	if ce == nil {
		return errors.New("lite: nil clock event")
	}

	inst.mu.RLock()
	listeners := append([]ClockListener(nil), inst.listeners...)
	inst.mu.RUnlock()

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

////////////////////////////////////////////////////////////////////////////////
// EOF
