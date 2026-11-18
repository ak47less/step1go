package lib

////////////////////////////////////////////////////////////////////////////////

type EventName string

type EventType int

type EventID int

type EventCallbackID int

////////////////////////////////////////////////////////////////////////////////

type Event struct {
	Name EventName

	Type EventType

	ID EventID

	Sender any

	Dispatcher EventDispatcher
}

////////////////////////////////////////////////////////////////////////////////

type EventDispatcher interface {
	Dispatch(e *Event) error

	AddListener(l EventListener) *EventCallback

	RemoveListener(ec *EventCallback)
}

////////////////////////////////////////////////////////////////////////////////

type EventListener interface {
	GetCallback() *EventCallback
}

////////////////////////////////////////////////////////////////////////////////

type EventCallback struct {
	Priority int

	ID EventCallbackID

	Name EventName

	Type EventType

	Handler func(e *Event) error
}

////////////////////////////////////////////////////////////////////////////////

type DefaultEventDispatcher struct {
}

// AddListener implements [EventDispatcher].
func (inst *DefaultEventDispatcher) AddListener(l EventListener) *EventCallback {
	panic("unimplemented")
}

// Dispatch implements [EventDispatcher].
func (inst *DefaultEventDispatcher) Dispatch(e *Event) error {
	panic("unimplemented")
}

// RemoveListener implements [EventDispatcher].
func (inst *DefaultEventDispatcher) RemoveListener(ec *EventCallback) {
	panic("unimplemented")
}

// GetLifeCycle implements [Life].
func (inst *DefaultEventDispatcher) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit: inst.onInit,
	}
	return lc
}

func (inst *DefaultEventDispatcher) onInit(ac *ApplicationContext) error {
	ac.EventDispatcher = inst
	return nil
}

func (inst *DefaultEventDispatcher) _impl() (EventDispatcher, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
