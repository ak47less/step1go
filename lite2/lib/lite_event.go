package lib

type EventType int

type Event struct {
	Type EventType
}

type EventDispatcher interface {
}

type EventListener struct {
	Callback func(e *Event) error
}
