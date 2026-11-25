package lib

////////////////////////////////////////////////////////////////////////////////

const (
	EventTypeMin EventType = iota

	EventTypeOnAddTrack
	EventTypeOnRemoveTrack
	EventTypeOnChangeBank

	EventTypeMax
)

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
	context *ApplicationContext
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
		OnInit:   inst.onInit,
		OnCreate: inst.onCreate,
	}
	return lc
}

func (inst *DefaultEventDispatcher) onInit(ac *ApplicationContext) error {
	ac.EventDispatcher = inst
	inst.context = ac
	return nil
}

func (inst *DefaultEventDispatcher) onCreate() error {

	ob := &ClockObserver{
		Preload: false,
		Unit:    ClockUnitFrame,
		Handler: inst.onFrameClock,
	}

	inst.context.Clock.AddObserver(ob)

	return nil
}

func (inst *DefaultEventDispatcher) onFrameClock(e *ClockEvent) error {

	logger := inst.context.Logger

	const f = "DefaultEventDispatcher: on_frame() [event index:%d unit:%d unit_len:%f realtime:%f thistime:%f ]"
	logger.Info(f, e.Index, e.Unit, e.UnitLength, e.Realtime, e.Thistime)

	return nil
}

func (inst *DefaultEventDispatcher) _impl() (EventDispatcher, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
