package lite

type EventType int

type EventID int

////////////////////////////////////////////////////////////////////////////////

const (
	EventTypeMin EventType = iota

	EventTypeMax
)

////////////////////////////////////////////////////////////////////////////////

type Event struct {
	Type    EventType
	ID      EventID
	Sender  any
	Context *Context
}

type EventCallbackFunc func(e *Event) error

type EventCallback struct {
	Type EventType
	ID   EventID
	Fn   EventCallbackFunc
}

type EventDispatcher interface {
	Dispatch(e Event) error

	AddCallback(c *EventCallback) error

	RemoveCallback(c *EventCallback) error
}

////////////////////////////////////////////////////////////////////////////////
