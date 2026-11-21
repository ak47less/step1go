package lib

type Transport interface {
	Play() error
	Record() error
	Stop() error
	Pause() error
	Resume() error

	Seek(to *RelativePosition) error

	GetCurrentPosition() *RelativePosition
}

type TransportContext struct {
	AC        *ApplicationContext
	StopAt    *RelativePosition
	Transport Transport

	Playing   bool
	Recording bool
	Paused    bool
}

////////////////////////////////////////////////////////////////////////////////

func NewTransport() Transport {
	return new(innerTransport)
}

////////////////////////////////////////////////////////////////////////////////

type innerTransport struct {
}

// GetCurrentPosition implements [Transport].
func (inst *innerTransport) GetCurrentPosition() *RelativePosition {
	panic("unimplemented")
}

// Pause implements [Transport].
func (inst *innerTransport) Pause() error {
	panic("unimplemented")
}

// Play implements [Transport].
func (inst *innerTransport) Play() error {
	panic("unimplemented")
}

// Record implements [Transport].
func (inst *innerTransport) Record() error {
	panic("unimplemented")
}

// Resume implements [Transport].
func (inst *innerTransport) Resume() error {
	panic("unimplemented")
}

// Seek implements [Transport].
func (inst *innerTransport) Seek(to *RelativePosition) error {
	panic("unimplemented")
}

// Stop implements [Transport].
func (inst *innerTransport) Stop() error {
	panic("unimplemented")
}

// GetLifeCycle implements [Life].
func (inst *innerTransport) GetLifeCycle() *LifeCycle {
	return &LifeCycle{
		Order: 0,
	}
}

func (inst *innerTransport) innerOnInit(ac *ApplicationContext) error {

	ac.TransportContext.Transport = inst
	ac.Transport = inst

	return nil
}

func (inst *innerTransport) _impl() (Transport, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
