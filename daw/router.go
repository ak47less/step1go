package step1core

type Router interface {
	OnDevice(d Device) error

	OnPort(p Port) error

	OnWire(w *Wire) error
}
